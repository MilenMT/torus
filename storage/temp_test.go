package storage

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/alternative-storage/torus"
)

const nblocks = 10

func TestTempWriteBlock(t *testing.T) {
	tbs := &tempBlockStore{
		store:     make(map[torus.BlockRef][]byte),
		nBlocks:   nblocks,
		blockSize: BlockSize,
	}

	if tbs.Kind() != "temp" {
		t.Fatalf("Got wrong kind %s, expected %s", tbs.Kind(), "temp")
	}

	data := make([]byte, 256)
	order := binary.LittleEndian

	order.PutUint64(data, 0x0101010101010101)
	ref := torus.BlockRefFromBytes(data)

	err := tbs.WriteBlock(nil, ref, data)
	if err != nil {
		t.Fatal(err)
	}

	err = tbs.Flush()
	if err != nil {
		t.Fatal(err)
	}

	ret, err := tbs.GetBlock(nil, ref)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(ret[:len(data)], data) {
		t.Fatal("Got useless data back")
	}

	// test HasBlock
	has, err := tbs.HasBlock(nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal("Data disappeared")
	}

	// test NumBlocks
	if tbs.NumBlocks() != nblocks {
		t.Fatalf("Wrong number of blocks %d, expected %d", tbs.NumBlocks(), nblocks)
	}

	// test UsedBlocks
	if tbs.UsedBlocks() != 1 {
		t.Fatalf("Wrong number of used blocks %d, expected %d", tbs.UsedBlocks(), 1)
	}

	// test DeleteBlock
	err = tbs.DeleteBlock(nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	has, err = tbs.HasBlock(nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Fatal("Failed to delete")
	}

	// test Close
	err = tbs.Close()
	if err != nil {
		t.Fatal(err)
	}
	if tbs.store != nil {
		t.Fatal("Failed to close")
	}
}
