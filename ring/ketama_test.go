package ring

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alternative-storage/torus"
	"github.com/alternative-storage/torus/models"
	"github.com/serialx/hashring"
)

func makeTinyPeers(t *testing.T) torus.Ring {
	pi := torus.PeerInfoList{
		&models.PeerInfo{
			UUID:        "a",
			TotalBlocks: 20 * 1024 * 1024 * 2,
		},
		&models.PeerInfo{
			UUID:        "b",
			TotalBlocks: 20 * 1024 * 1024 * 2,
		},
		&models.PeerInfo{
			UUID:        "c",
			TotalBlocks: 100 * 1024 * 2,
		},
	}
	return &ketama{
		version: 1,
		peers:   pi,
		rep:     2,
		ring:    hashring.NewWithWeights(pi.GetWeights()),
	}
}

func TestGetPeer(t *testing.T) {
	k := makeTinyPeers(t)
	l, err := k.GetPeers(torus.BlockRef{
		INodeRef: torus.NewINodeRef(3, 4),
		Index:    5,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(l.Peers)
}

func TestAddPeers(t *testing.T) {
	newpi := torus.PeerInfoList{
		&models.PeerInfo{
			UUID:        "d",
			TotalBlocks: 20 * 1024 * 1024 * 2,
		},
	}
	k := makeTinyPeers(t).(torus.RingAdder)
	newk, err := k.AddPeers(newpi)
	if err != nil {
		t.Fatal(err)
	}
	if newk.Version() != 2 {
		t.Fatalf("Faild to update version to %d, expected %d", newk.Version(), 2)
	}
}

func TestRemovePeers(t *testing.T) {
	k := makeTinyPeers(t).(torus.RingRemover)
	rmpi := torus.PeerList{"a"}
	newk, err := k.RemovePeers(rmpi)
	if err != nil {
		t.Fatal(err)
	}
	if newk.Version() != 2 {
		t.Fatalf("Faild to update version to %d, expected %d", newk.Version(), 2)
	}
	pl := newk.Members()
	if len(pl) != 2 {
		t.Fatalf("Faild to remove peer member")
	}
}

func TestChangeReplication(t *testing.T) {
	newRep := 3
	k := makeTinyPeers(t).(torus.ModifyableRing)
	newk, err := k.ChangeReplication(newRep)
	if err != nil {
		t.Fatal(err)
	}
	if newk.Version() != 2 {
		t.Fatalf("Faild to update version to %d, expected %d", newk.Version(), 2)
	}
	if !strings.Contains(newk.Describe(), fmt.Sprintf("Ring: Ketama\nReplication:%d\nPeers:", newRep)) {
		t.Fatalf("Got wrong replications")
	}
}
