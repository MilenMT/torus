package distributor

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/net/context"
	"net/url"
	"testing"

	"github.com/alternative-storage/torus"
)

func TestWriteBlock(t *testing.T) {
	t.Log("Write and Read Blocks")
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
	defer dist.Close()

	data := make([]byte, BlockSize)
	order := binary.LittleEndian

	order.PutUint64(data, 0x0101010101010101)
	ref := torus.BlockRefFromBytes(data)

	err = dist.WriteBlock(context.Background(), ref, data)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := dist.GetBlock(context.Background(), ref)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(ret, data) {
		t.Fatalf("read operation didn't return expected data for block ref %q\nexpected first 8 bytes: %v\nactual first 8 bytes: %v", data, ret[:8], data[:8])
	}
}
