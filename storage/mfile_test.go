package storage

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/alternative-storage/torus"
)

const (
	BlockSize  = 1024
	mBlockSize = torus.BlockRefByteSize
)

func TestWriteBlock(t *testing.T) {
	mpath, err := makeTempFilename()
	if err != nil {
		t.Fatal(err)
	}
	m, err := CreateOrOpenMFile(mpath, mBlockSize*10, mBlockSize)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	dpath, err := makeTempFilename()
	if err != nil {
		t.Fatal(err)
	}
	d, err := CreateOrOpenMFile(dpath, BlockSize*10, BlockSize)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	mfb := &mfileBlock{
		refFile:  m,
		dataFile: d,
	}
	refIndex, err := loadIndex(m)
	if err != nil {
		t.Fatal(err)
	}
	mfb.refIndex = refIndex

	data := make([]byte, 256)
	order := binary.LittleEndian

	order.PutUint64(data, 0x0101010101010101)
	ref := torus.BlockRefFromBytes(data)

	err = mfb.WriteBlock(nil, ref, data)
	if err != nil {
		t.Fatal(err)
	}

	mfb.Flush()
	ret, err := mfb.GetBlock(nil, ref)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(ret[:len(data)], data) {
		t.Fatal("Got useless data back")
	}

	// test HasBlock
	has, err := mfb.HasBlock(nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal("Data disappeared")
	}

	// test DeleteBlock
	err = mfb.DeleteBlock(nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	has, err = mfb.HasBlock(nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Fatal("Failed to delete")
	}
}
