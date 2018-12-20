package block

import (
	"testing"

	"github.com/alternative-storage/torus"
	"github.com/alternative-storage/torus/metadata/temp"

	// CreateBlockStore
	_ "github.com/alternative-storage/torus/storage"
)

func TestOpenBlockFile(t *testing.T) {
	// Create
	md := temp.NewServer()
	cfg := torus.Config{
		StorageSize: 100 * 1024 * 1024,
	}
	mds := temp.NewClient(cfg, md)
	err := CreateBlockVolume(mds, volName, 1024)
	if err != nil {
		t.Fatal(err)
	}

	// Open
	srv := newServer(md)
	openvol, err := OpenBlockVolume(srv, volName)
	if err != nil {
		t.Fatal(err)
	}

	// OpenBlockFile
	bf, err := openvol.OpenBlockFile()
	if err != nil {
		t.Fatal(err)
	}
	if bf.vol != openvol {
		t.Fatal("Got wrong volume")
	}

	// SaveSnapshot
	err = openvol.SaveSnapshot(snapName)
	if err != nil {
		t.Fatal(err)
	}

	// OpenSnapshot
	fromsnap, err := openvol.OpenSnapshot(snapName)
	if err != nil {
		t.Fatal(err)
	}
	if fromsnap.vol != openvol {
		t.Fatal("Got wrong volume")
	}

	// CloseBlockFile
	err = bf.Close()
	if err != nil {
		t.Fatal(err)
	}

	// TODO: should modify original openvol before restore
	restoreVol := openvol
	// RestoreSnapshot
	err = restoreVol.RestoreSnapshot(snapName)
	if err != nil {
		t.Fatal(err)
	}
	if restoreVol != openvol {
		t.Fatal("Got wrong volume")
	}
}
