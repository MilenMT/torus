package distributor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/net/context"
	"math/rand"
	"net/url"
	"testing"
	"time"

	"github.com/alternative-storage/torus"
	"github.com/alternative-storage/torus/metadata/temp"
	"github.com/alternative-storage/torus/models"
	"github.com/alternative-storage/torus/ring"
)

const (
	StorageSize = 512 * 1024
	BlockSize   = 256
)

func TestPutBlock(t *testing.T) {
	t.Log("Write and Read through gRPC")
	srvs, _ := ringN(t, 3)
	defer func() {
		for _, s := range srvs {
			s.Close()
		}
	}()
	p, err := srvs[0].MDS.GetPeers()
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(p))
	}
	addr := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:0",
	}

	dist, err := newDistributor(srvs[0], addr)
	defer dist.Close()
	if err != nil {
		t.Fatal(err)
	}

	data := make([]byte, BlockSize)
	order := binary.LittleEndian

	order.PutUint64(data, 0x0101010101010101)
	ref := torus.BlockRefFromBytes(data)

	err = dist.PutBlock(context.Background(), ref, data)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := dist.Block(context.Background(), ref)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(ret, data) {
		t.Fatalf("read operation didn't return expected data for block ref %q\nexpected first 8 bytes: %v\nactual first 8 bytes: %v", data, ret[:8], data[:8])
	}
}

func createN(t testing.TB, n int) ([]*torus.Server, *temp.Server) {
	var out []*torus.Server
	s := temp.NewServer()
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		srv := newServer(s)
		// TODO :0 doesn't work and currently uses workaround.
		addr := fmt.Sprintf("http://127.0.0.1:%d", 20000+rand.Intn(1000))
		uri, err := url.Parse(addr)
		if err != nil {
			t.Fatal(err)
		}
		err = ListenReplication(srv, uri)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, srv)
	}
	// Heartbeat
	time.Sleep(10 * time.Millisecond)
	return out, s
}

func ringN(t testing.TB, n int) ([]*torus.Server, *temp.Server) {
	servers, mds := createN(t, n)
	var peers torus.PeerInfoList
	for _, s := range servers {
		peers = append(peers, &models.PeerInfo{
			UUID:        s.MDS.UUID(),
			TotalBlocks: StorageSize / BlockSize,
		})
	}

	rep := 2
	ringType := ring.Ketama
	if n == 1 {
		rep = 1
		ringType = ring.Single
	}

	newRing, err := ring.CreateRing(&models.Ring{
		Type:              uint32(ringType),
		Peers:             peers,
		ReplicationFactor: uint32(rep),
		Version:           uint32(2),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = mds.SetRing(newRing)
	if err != nil {
		t.Fatal(err)
	}
	// wait here, otherwise would get "no peers available for a block"
	time.Sleep(10 * time.Millisecond)
	return servers, mds
}
