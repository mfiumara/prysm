package kv

import (
	"context"
	"flag"
	"reflect"
	"testing"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/shared/params"
	"gopkg.in/urfave/cli.v2"
)

func TestNilDBHistoryBlkHdr(t *testing.T) {
	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	db := setupDB(t, cli.NewContext(&app, set, nil))
	defer teardownDB(t, db)
	ctx := context.Background()

	epoch := uint64(1)
	validatorID := uint64(1)

	hasBlockHeader := db.HasBlockHeader(ctx, epoch, validatorID)
	if hasBlockHeader {
		t.Fatal("HasBlockHeader should return false")
	}

	bPrime, err := db.BlockHeaders(ctx, epoch, validatorID)
	if err != nil {
		t.Fatalf("failed to get block: %v", err)
	}
	if bPrime != nil {
		t.Fatalf("get should return nil for a non existent key")
	}
}

func TestSaveHistoryBlkHdr(t *testing.T) {
	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	db := setupDB(t, cli.NewContext(&app, set, nil))
	ctx := context.Background()

	tests := []struct {
		vID uint64
		bh  *ethpb.SignedBeaconBlockHeader
	}{
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in"), Header: &ethpb.BeaconBlockHeader{Slot: 0}},
		},
		{
			vID: uint64(1),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 2nd"), Header: &ethpb.BeaconBlockHeader{Slot: 0}},
		},
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 3rd"), Header: &ethpb.BeaconBlockHeader{Slot: params.BeaconConfig().SlotsPerEpoch + 1}},
		},
	}

	for _, tt := range tests {
		err := db.SaveBlockHeader(ctx, tt.vID, tt.bh)
		if err != nil {
			t.Fatalf("save block failed: %v", err)
		}

		bha, err := db.BlockHeaders(ctx, helpers.SlotToEpoch(tt.bh.Header.Slot), tt.vID)
		if err != nil {
			t.Fatalf("failed to get block: %v", err)
		}

		if bha == nil || !reflect.DeepEqual(bha[0], tt.bh) {
			t.Fatalf("get should return bh: %v", bha)
		}
	}

}

func TestDeleteHistoryBlkHdr(t *testing.T) {
	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	db := setupDB(t, cli.NewContext(&app, set, nil))
	defer teardownDB(t, db)
	ctx := context.Background()

	tests := []struct {
		vID uint64
		bh  *ethpb.SignedBeaconBlockHeader
	}{
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in"), Header: &ethpb.BeaconBlockHeader{Slot: 0}},
		},
		{
			vID: uint64(1),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 2nd"), Header: &ethpb.BeaconBlockHeader{Slot: 0}},
		},
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 3rd"), Header: &ethpb.BeaconBlockHeader{Slot: params.BeaconConfig().SlotsPerEpoch + 1}},
		},
	}
	for _, tt := range tests {

		err := db.SaveBlockHeader(ctx, tt.vID, tt.bh)
		if err != nil {
			t.Fatalf("save block failed: %v", err)
		}
	}

	for _, tt := range tests {
		bha, err := db.BlockHeaders(ctx, helpers.SlotToEpoch(tt.bh.Header.Slot), tt.vID)
		if err != nil {
			t.Fatalf("failed to get block: %v", err)
		}

		if bha == nil || !reflect.DeepEqual(bha[0], tt.bh) {
			t.Fatalf("get should return bh: %v", bha)
		}
		err = db.DeleteBlockHeader(ctx, tt.vID, tt.bh)
		if err != nil {
			t.Fatalf("save block failed: %v", err)
		}
		bh, err := db.BlockHeaders(ctx, helpers.SlotToEpoch(tt.bh.Header.Slot), tt.vID)

		if err != nil {
			t.Fatal(err)
		}
		if bh != nil {
			t.Errorf("Expected block to have been deleted, received: %v", bh)
		}

	}

}

func TestHasHistoryBlkHdr(t *testing.T) {
	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	db := setupDB(t, cli.NewContext(&app, set, nil))
	defer teardownDB(t, db)
	ctx := context.Background()

	tests := []struct {
		vID uint64
		bh  *ethpb.SignedBeaconBlockHeader
	}{
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in"), Header: &ethpb.BeaconBlockHeader{Slot: 0}},
		},
		{
			vID: uint64(1),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 2nd"), Header: &ethpb.BeaconBlockHeader{Slot: 0}},
		},
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 3rd"), Header: &ethpb.BeaconBlockHeader{Slot: params.BeaconConfig().SlotsPerEpoch + 1}},
		},
	}
	for _, tt := range tests {

		found := db.HasBlockHeader(ctx, helpers.SlotToEpoch(tt.bh.Header.Slot), tt.vID)
		if found {
			t.Fatal("has block header should return false for block headers that are not in db")
		}
		err := db.SaveBlockHeader(ctx, tt.vID, tt.bh)
		if err != nil {
			t.Fatalf("save block failed: %v", err)
		}
	}
	for _, tt := range tests {
		err := db.SaveBlockHeader(ctx, tt.vID, tt.bh)
		if err != nil {
			t.Fatalf("save block failed: %v", err)
		}

		found := db.HasBlockHeader(ctx, helpers.SlotToEpoch(tt.bh.Header.Slot), tt.vID)

		if !found {
			t.Fatal("has block header should return true")
		}
	}
}

func TestPruneHistoryBlkHdr(t *testing.T) {
	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	db := setupDB(t, cli.NewContext(&app, set, nil))
	defer teardownDB(t, db)
	ctx := context.Background()

	tests := []struct {
		vID uint64
		bh  *ethpb.SignedBeaconBlockHeader
	}{
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in"), Header: &ethpb.BeaconBlockHeader{Slot: 0}},
		},
		{
			vID: uint64(1),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 2nd"), Header: &ethpb.BeaconBlockHeader{Slot: 0}},
		},
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 3rd"), Header: &ethpb.BeaconBlockHeader{Slot: params.BeaconConfig().SlotsPerEpoch + 1}},
		},
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 4th"), Header: &ethpb.BeaconBlockHeader{Slot: params.BeaconConfig().SlotsPerEpoch*2 + 1}},
		},
		{
			vID: uint64(0),
			bh:  &ethpb.SignedBeaconBlockHeader{Signature: []byte("let me in 5th"), Header: &ethpb.BeaconBlockHeader{Slot: params.BeaconConfig().SlotsPerEpoch*3 + 1}},
		},
	}

	for _, tt := range tests {
		err := db.SaveBlockHeader(ctx, tt.vID, tt.bh)
		if err != nil {
			t.Fatalf("save block header failed: %v", err)
		}

		bha, err := db.BlockHeaders(ctx, helpers.SlotToEpoch(tt.bh.Header.Slot), tt.vID)
		if err != nil {
			t.Fatalf("failed to get block header: %v", err)
		}

		if bha == nil || !reflect.DeepEqual(bha[0], tt.bh) {
			t.Fatalf("get should return bh: %v", bha)
		}
	}
	currentEpoch := uint64(3)
	historyToKeep := uint64(2)
	err := db.PruneBlockHistory(ctx, currentEpoch, historyToKeep)
	if err != nil {
		t.Fatalf("failed to prune: %v", err)
	}

	for _, tt := range tests {
		bha, err := db.BlockHeaders(ctx, helpers.SlotToEpoch(tt.bh.Header.Slot), tt.vID)
		if err != nil {
			t.Fatalf("failed to get block header: %v", err)
		}
		if helpers.SlotToEpoch(tt.bh.Header.Slot) > currentEpoch-historyToKeep {
			if bha == nil || !reflect.DeepEqual(bha[0], tt.bh) {
				t.Fatalf("get should return bh: %v", bha)
			}
		} else {
			if bha != nil {
				t.Fatalf("block header should have been pruned: %v", bha)
			}
		}
	}
}
