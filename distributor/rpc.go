package distributor

import (
	"github.com/alternative-storage/torus"

	"github.com/coreos/pkg/capnslog"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
)

func (d *Distributor) Block(ctx context.Context, ref torus.BlockRef) ([]byte, error) {
	tracer := opentracing.GlobalTracer()
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span := tracer.StartSpan("Reading from storage", opentracing.ChildOf(span.Context()))
		span.SetTag("INode", ref.INodeRef.String())
		ctx = opentracing.ContextWithSpan(ctx, span)
		defer span.Finish()
	} else {
		clog.Warningf("failed to create span for Block")
	}
	promDistBlockRPCs.Inc()
	data, err := d.blocks.GetBlock(ctx, ref)
	if err != nil {
		promDistBlockRPCFailures.Inc()
		clog.Warningf("remote asking for non-existent block: %s", ref)
		return nil, torus.ErrBlockUnavailable
	}
	if torus.BlockLog.LevelAt(capnslog.TRACE) {
		torus.BlockLog.Tracef("rpc: retrieved block %s", ref)
	}
	return data, nil
}

// PutBlock server side implementation which is called from RPC client.
func (d *Distributor) PutBlock(ctx context.Context, ref torus.BlockRef, data []byte) error {
	tracer := opentracing.GlobalTracer()
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span := tracer.StartSpan("Writing to storage", opentracing.ChildOf(span.Context()))
		span.SetTag("INode", ref.INodeRef.String())
		ctx = opentracing.ContextWithSpan(ctx, span)
		defer span.Finish()
	} else {
		clog.Warningf("failed to create span for PutBlock")
	}
	d.mut.RLock()
	defer d.mut.RUnlock()
	promDistPutBlockRPCs.Inc()
	// get peer list belong to the blockref.
	peers, err := d.ring.GetPeers(ref)
	if err != nil {
		promDistPutBlockRPCFailures.Inc()
		return err
	}
	ok := false
	for _, x := range peers.Peers {
		// check if the UUID is same to me.
		if x == d.UUID() {
			ok = true
			break
		}
	}
	if !ok {
		clog.Warningf("trying to write block that doesn't belong to me.")
	}

	// WriteBlock to my storage file.
	err = d.blocks.WriteBlock(ctx, ref, data)
	if err != nil {
		return err
	}
	if torus.BlockLog.LevelAt(capnslog.TRACE) {
		torus.BlockLog.Tracef("rpc: saving block %s", ref)
	}
	return d.Flush()
}

func (d *Distributor) RebalanceCheck(ctx context.Context, refs []torus.BlockRef) ([]bool, error) {
	tracer := opentracing.GlobalTracer()
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span := tracer.StartSpan("Checking rebalnce", opentracing.ChildOf(span.Context()))
		// TODO set tag
		ctx = opentracing.ContextWithSpan(ctx, span)
		defer span.Finish()
	} else {
		clog.Warningf("failed to create span for RebalanceCheck")
	}
	out := make([]bool, len(refs))
	for i, x := range refs {
		ok, err := d.blocks.HasBlock(ctx, x)
		if err != nil {
			clog.Error(err)
			return nil, err
		}
		out[i] = ok
	}
	return out, nil
}
