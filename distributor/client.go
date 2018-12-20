package distributor

import (
	"net/url"
	//"runtime/debug"
	"sync"
	"time"

	"github.com/alternative-storage/torus"
	"github.com/alternative-storage/torus/distributor/protocols"
	"github.com/opentracing/opentracing-go"
	//"github.com/opentracing/opentracing-go/log"

	"golang.org/x/net/context"
)

const (
	connectTimeout         = 2 * time.Second
	rebalanceClientTimeout = 5 * time.Second
	clientTimeout          = 500 * time.Millisecond
	writeClientTimeout     = 2000 * time.Millisecond
)

// TODO(barakmich): Clean up errors

// distClient used by client side as a Distributor with connection.
type distClient struct {
	dist *Distributor
	//TODO(barakmich): Better connection pooling
	openConns map[string]protocols.RPC
	mut       sync.Mutex

	tracer opentracing.Tracer
}

// newDistClient returns distClient.
func newDistClient(d *Distributor) *distClient {
	client := &distClient{
		dist:      d,
		openConns: make(map[string]protocols.RPC),
	}
	d.srv.AddTimeoutCallback(client.onPeerTimeout)
	return client
}

func (d *distClient) onPeerTimeout(uuid string) {
	d.mut.Lock()
	defer d.mut.Unlock()
	conn, ok := d.openConns[uuid]
	if !ok {
		return
	}
	err := conn.Close()
	if err != nil {
		clog.Errorf("peer timeout err on close: %s", err)
	}
	delete(d.openConns, uuid)

}

// getConn returns RPC interface after found proper peers.
func (d *distClient) getConn(uuid string) protocols.RPC {
	d.mut.Lock()
	if conn, ok := d.openConns[uuid]; ok {
		d.mut.Unlock()
		return conn
	}
	d.mut.Unlock()
	pm := d.dist.srv.GetPeerMap()
	pi := pm[uuid]
	if pi == nil {
		// We know this UUID exists, we don't have an address for it, let's refresh now.
		pm := d.dist.srv.UpdatePeerMap()
		pi = pm[uuid]
		if pi == nil {
			// Not much more we can try
			return nil
		}
	}
	if pi.TimedOut {
		return nil
	}
	uri, err := url.Parse(pi.Address)
	if err != nil {
		clog.Errorf("couldn't parse address %s: %v", pi.Address, err)
		return nil
	}
	gmd := d.dist.srv.MDS.GlobalMetadata()
	conn, err := protocols.DialRPC(uri, connectTimeout, gmd)
	d.mut.Lock()
	defer d.mut.Unlock()
	if err != nil {
		clog.Errorf("couldn't dial: %v", err)
		return nil
	}
	d.openConns[uuid] = conn
	return conn
}

// resetConn resets connection for the uuid.
func (d *distClient) resetConn(uuid string) {
	d.mut.Lock()
	defer d.mut.Unlock()
	conn, ok := d.openConns[uuid]
	if !ok {
		return
	}
	delete(d.openConns, uuid)
	err := conn.Close()
	if err != nil {
		clog.Errorf("error resetConn: %s", err)
	}
}

// Close closes all opened connections.
func (d *distClient) Close() error {
	d.mut.Lock()
	defer d.mut.Unlock()
	for _, c := range d.openConns {
		err := c.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetBlock starts RPC call to get data.
func (d *distClient) GetBlock(ctx context.Context, uuid string, b torus.BlockRef) ([]byte, error) {
	conn := d.getConn(uuid)
	if conn == nil {
		return nil, torus.ErrNoPeer
	}
	data, err := conn.Block(ctx, b)
	if err != nil {
		d.resetConn(uuid)
		clog.Debug(err)
		return nil, torus.ErrBlockUnavailable
	}
	return data, nil
}

// PutBlock starts RPC call to put data.
func (d *distClient) PutBlock(ctx context.Context, uuid string, b torus.BlockRef, data []byte) error {
	conn := d.getConn(uuid)
	if conn == nil {
		return torus.ErrNoPeer
	}
	err := conn.PutBlock(ctx, b, data)
	if err != nil {
		d.resetConn(uuid)
		if err == context.DeadlineExceeded {
			return torus.ErrBlockUnavailable
		}
	}
	return err
}

func (d *distClient) Check(ctx context.Context, uuid string, blks []torus.BlockRef) ([]bool, error) {
	conn := d.getConn(uuid)
	if conn == nil {
		return nil, torus.ErrNoPeer
	}
	resp, err := conn.RebalanceCheck(ctx, blks)
	if err != nil {
		d.resetConn(uuid)
		return nil, err
	}
	return resp, nil
}
