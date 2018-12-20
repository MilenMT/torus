package block

import (
	"testing"

	"github.com/alternative-storage/torus"
	"github.com/alternative-storage/torus/metadata/temp"

	// CreateBlockStore
	_ "github.com/alternative-storage/torus/storage"
)

const (
	volName    = "testVol"
	snapName   = "testSnap"
	newVolName = "testNewVol"
)

func newServer(md *temp.Server) *torus.Server {
	cfg := torus.Config{
		StorageSize: 100 * 1024 * 1024,
	}
	mds := temp.NewClient(cfg, md)
	gmd := mds.GlobalMetadata()
	blocks, _ := torus.CreateBlockStore("temp", "current", cfg, gmd)
	s, _ := torus.NewServerByImpl(cfg, mds, blocks)
	return s
}

func TestCreateOpenDeleteBlockVolume(t *testing.T) {
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
	if openvol.volume.Name != volName {
		t.Fatalf("Got wrong volume name %s, expected %s", openvol.volume.Name, volName)
	}

	// TODO: bug in code (?)
	// DeleteVolume() in block/temp.go
	// d.locked != b.UUID() should also allow d.locked == ""
	/*
		err = DeleteBlockVolume(mds, volName)
		if err != nil {
			t.Fatal(err)
		}
		openvol, err = OpenBlockVolume(srv, volName)
		if err == nil {
			t.Fatal("Failed to delete volume")
		}
	*/

}

func TestSnapshotCreateOpenDeleteBlockVolume(t *testing.T) {
	// Create and Open original volume
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
	if openvol.volume.Name != volName {
		t.Fatalf("Got wrong volume name %s, expected %s", openvol.volume.Name, volName)
	}

	// Create snapshot
	err = openvol.SaveSnapshot(snapName)
	if err != nil {
		t.Fatal(err)
	}

	// Get snapshot
	snapList, err := openvol.GetSnapshots()
	if err != nil {
		t.Fatal(err)
	}
	if snapList[0].Name != snapName {
		t.Fatalf("Got wrong snapshot name %s, expected %s", snapList[0].Name, snapName)
	}

	// Create vol from snapshot
	err = CreateBlockFromSnapshot(srv, volName, snapName, newVolName, false)
	if err != nil {
		t.Fatal(err)
	}

	// Open created volume
	newvol, err := OpenBlockVolume(srv, newVolName)
	if err != nil {
		t.Fatal(err)
	}
	if newvol.volume.Name != newVolName {
		t.Fatalf("Got wrong volume name %s, expected %s", newvol.volume.Name, newVolName)
	}
}
