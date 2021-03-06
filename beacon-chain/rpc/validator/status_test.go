package validator

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-ssz"
	mockChain "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/cache/depositcache"
	blk "github.com/prysmaticlabs/prysm/beacon-chain/core/blocks"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	dbutil "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	mockPOW "github.com/prysmaticlabs/prysm/beacon-chain/powchain/testing"
	stateTrie "github.com/prysmaticlabs/prysm/beacon-chain/state"
	pbp2p "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/trieutil"
)

func TestValidatorStatus_Deposited(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	ctx := context.Background()

	pubKey := pubKey(1)
	depData := &ethpb.Deposit_Data{
		PublicKey:             pubKey,
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
	}
	deposit := &ethpb.Deposit{
		Data: depData,
	}
	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()
	depositCache.InsertDeposit(ctx, deposit, 0 /*blockNum*/, 0, depositTrie.Root())
	height := time.Unix(int64(params.BeaconConfig().Eth1FollowDistance), 0).Unix()
	p := &mockPOW.POWChain{
		TimesByHeight: map[int]uint64{
			0: uint64(height),
		},
	}
	stateObj, err := stateTrie.InitializeFromProtoUnsafe(&pbp2p.BeaconState{})
	if err != nil {
		t.Fatal(err)
	}
	vs := &Server{
		BeaconDB:       db,
		DepositFetcher: depositCache,
		BlockFetcher:   p,
		HeadFetcher: &mockChain.ChainService{
			State: stateObj,
		},
		Eth1InfoFetcher: p,
	}
	req := &ethpb.ValidatorStatusRequest{
		PublicKey: pubKey,
	}
	resp, err := vs.ValidatorStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("Could not get validator status %v", err)
	}
	if resp.Status != ethpb.ValidatorStatus_DEPOSITED {
		t.Errorf("Wanted %v, got %v", ethpb.ValidatorStatus_DEPOSITED, resp.Status)
	}
}

func TestValidatorStatus_Pending(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	ctx := context.Background()

	pubKey := pubKey(1)
	if err := db.SaveValidatorIndex(ctx, pubKey, 0); err != nil {
		t.Fatalf("Could not save validator index: %v", err)
	}
	block := blk.NewGenesisBlock([]byte{})
	if err := db.SaveBlock(ctx, block); err != nil {
		t.Fatalf("Could not save genesis block: %v", err)
	}
	genesisRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		t.Fatalf("Could not get signing root %v", err)
	}
	// Pending active because activation epoch is still defaulted at far future slot.
	state, _ := stateTrie.InitializeFromProto(&pbp2p.BeaconState{
		Validators: []*ethpb.Validator{
			{
				ActivationEpoch:   params.BeaconConfig().FarFutureEpoch,
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
				PublicKey:         pubKey,
			},
		},
		Slot: 5000,
	})
	if err := db.SaveState(ctx, state, genesisRoot); err != nil {
		t.Fatalf("could not save state: %v", err)
	}
	if err := db.SaveHeadBlockRoot(ctx, genesisRoot); err != nil {
		t.Fatalf("Could not save genesis state: %v", err)
	}

	depData := &ethpb.Deposit_Data{
		PublicKey:             pubKey,
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
	}

	deposit := &ethpb.Deposit{
		Data: depData,
	}
	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()
	depositCache.InsertDeposit(ctx, deposit, 0 /*blockNum*/, 0, depositTrie.Root())

	height := time.Unix(int64(params.BeaconConfig().Eth1FollowDistance), 0).Unix()
	p := &mockPOW.POWChain{
		TimesByHeight: map[int]uint64{
			0: uint64(height),
		},
	}
	vs := &Server{
		BeaconDB:          db,
		ChainStartFetcher: p,
		BlockFetcher:      p,
		Eth1InfoFetcher:   p,
		DepositFetcher:    depositCache,
		HeadFetcher:       &mockChain.ChainService{State: state, Root: genesisRoot[:]},
	}
	req := &ethpb.ValidatorStatusRequest{
		PublicKey: pubKey,
	}
	resp, err := vs.ValidatorStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("Could not get validator status %v", err)
	}
	if resp.Status != ethpb.ValidatorStatus_PENDING {
		t.Errorf("Wanted %v, got %v", ethpb.ValidatorStatus_PENDING, resp.Status)
	}
}

func TestValidatorStatus_Active(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	// This test breaks if it doesnt use mainnet config
	params.OverrideBeaconConfig(params.MainnetConfig())
	defer params.OverrideBeaconConfig(params.MinimalSpecConfig())
	ctx := context.Background()

	pubKey := pubKey(1)
	if err := db.SaveValidatorIndex(ctx, pubKey, 0); err != nil {
		t.Fatalf("Could not save validator index: %v", err)
	}

	depData := &ethpb.Deposit_Data{
		PublicKey:             pubKey,
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
	}

	deposit := &ethpb.Deposit{
		Data: depData,
	}
	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()
	depositCache.InsertDeposit(ctx, deposit, 0 /*blockNum*/, 0, depositTrie.Root())

	// Active because activation epoch <= current epoch < exit epoch.
	activeEpoch := helpers.ActivationExitEpoch(0)

	block := blk.NewGenesisBlock([]byte{})
	if err := db.SaveBlock(ctx, block); err != nil {
		t.Fatalf("Could not save genesis block: %v", err)
	}
	genesisRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		t.Fatalf("Could not get signing root %v", err)
	}

	state := &pbp2p.BeaconState{
		GenesisTime: uint64(time.Unix(0, 0).Unix()),
		Slot:        10000,
		Validators: []*ethpb.Validator{{
			ActivationEpoch:   activeEpoch,
			ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
			WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			PublicKey:         pubKey},
		}}
	stateObj, err := stateTrie.InitializeFromProtoUnsafe(state)
	if err != nil {
		t.Fatal(err)
	}

	timestamp := time.Unix(int64(params.BeaconConfig().Eth1FollowDistance), 0).Unix()
	p := &mockPOW.POWChain{
		TimesByHeight: map[int]uint64{
			int(params.BeaconConfig().Eth1FollowDistance): uint64(timestamp),
		},
	}
	vs := &Server{
		BeaconDB:          db,
		ChainStartFetcher: p,
		BlockFetcher:      p,
		Eth1InfoFetcher:   p,
		DepositFetcher:    depositCache,
		HeadFetcher:       &mockChain.ChainService{State: stateObj, Root: genesisRoot[:]},
	}
	req := &ethpb.ValidatorStatusRequest{
		PublicKey: pubKey,
	}
	resp, err := vs.ValidatorStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("Could not get validator status %v", err)
	}

	expected := &ethpb.ValidatorStatusResponse{
		Status:               ethpb.ValidatorStatus_ACTIVE,
		ActivationEpoch:      5,
		DepositInclusionSlot: 2218,
	}
	if !proto.Equal(resp, expected) {
		t.Errorf("Wanted %v, got %v", expected, resp)
	}
}

func TestValidatorStatus_Exiting(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	ctx := context.Background()

	pubKey := pubKey(1)
	if err := db.SaveValidatorIndex(ctx, pubKey, 0); err != nil {
		t.Fatalf("Could not save validator index: %v", err)
	}

	// Initiated exit because validator exit epoch and withdrawable epoch are not FAR_FUTURE_EPOCH
	slot := uint64(10000)
	epoch := helpers.SlotToEpoch(slot)
	exitEpoch := helpers.ActivationExitEpoch(epoch)
	withdrawableEpoch := exitEpoch + params.BeaconConfig().MinValidatorWithdrawabilityDelay
	block := blk.NewGenesisBlock([]byte{})
	if err := db.SaveBlock(ctx, block); err != nil {
		t.Fatalf("Could not save genesis block: %v", err)
	}
	genesisRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		t.Fatalf("Could not get signing root %v", err)
	}

	state := &pbp2p.BeaconState{
		Slot: slot,
		Validators: []*ethpb.Validator{{
			PublicKey:         pubKey,
			ActivationEpoch:   0,
			ExitEpoch:         exitEpoch,
			WithdrawableEpoch: withdrawableEpoch},
		}}
	stateObj, err := stateTrie.InitializeFromProtoUnsafe(state)
	if err != nil {
		t.Fatal(err)
	}
	depData := &ethpb.Deposit_Data{
		PublicKey:             pubKey,
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
	}

	deposit := &ethpb.Deposit{
		Data: depData,
	}
	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()
	depositCache.InsertDeposit(ctx, deposit, 0 /*blockNum*/, 0, depositTrie.Root())
	height := time.Unix(int64(params.BeaconConfig().Eth1FollowDistance), 0).Unix()
	p := &mockPOW.POWChain{
		TimesByHeight: map[int]uint64{
			0: uint64(height),
		},
	}
	vs := &Server{
		BeaconDB:          db,
		ChainStartFetcher: p,
		BlockFetcher:      p,
		Eth1InfoFetcher:   p,
		DepositFetcher:    depositCache,
		HeadFetcher:       &mockChain.ChainService{State: stateObj, Root: genesisRoot[:]},
	}
	req := &ethpb.ValidatorStatusRequest{
		PublicKey: pubKey,
	}
	resp, err := vs.ValidatorStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("Could not get validator status %v", err)
	}
	if resp.Status != ethpb.ValidatorStatus_EXITING {
		t.Errorf("Wanted %v, got %v", ethpb.ValidatorStatus_EXITING, resp.Status)
	}
}

func TestValidatorStatus_Slashing(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	ctx := context.Background()

	pubKey := pubKey(1)
	if err := db.SaveValidatorIndex(ctx, pubKey, 0); err != nil {
		t.Fatalf("Could not save validator index: %v", err)
	}

	// Exit slashed because slashed is true, exit epoch is =< current epoch and withdrawable epoch > epoch .
	slot := uint64(10000)
	epoch := helpers.SlotToEpoch(slot)
	block := blk.NewGenesisBlock([]byte{})
	if err := db.SaveBlock(ctx, block); err != nil {
		t.Fatalf("Could not save genesis block: %v", err)
	}
	genesisRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		t.Fatalf("Could not get signing root %v", err)
	}

	state := &pbp2p.BeaconState{
		Slot: slot,
		Validators: []*ethpb.Validator{{
			Slashed:           true,
			PublicKey:         pubKey,
			WithdrawableEpoch: epoch + 1},
		}}
	stateObj, err := stateTrie.InitializeFromProtoUnsafe(state)
	if err != nil {
		t.Fatal(err)
	}
	depData := &ethpb.Deposit_Data{
		PublicKey:             pubKey,
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
	}

	deposit := &ethpb.Deposit{
		Data: depData,
	}
	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()
	depositCache.InsertDeposit(ctx, deposit, 0 /*blockNum*/, 0, depositTrie.Root())
	height := time.Unix(int64(params.BeaconConfig().Eth1FollowDistance), 0).Unix()
	p := &mockPOW.POWChain{
		TimesByHeight: map[int]uint64{
			0: uint64(height),
		},
	}
	vs := &Server{
		BeaconDB:          db,
		ChainStartFetcher: p,
		Eth1InfoFetcher:   p,
		DepositFetcher:    depositCache,
		BlockFetcher:      p,
		HeadFetcher:       &mockChain.ChainService{State: stateObj, Root: genesisRoot[:]},
	}
	req := &ethpb.ValidatorStatusRequest{
		PublicKey: pubKey,
	}
	resp, err := vs.ValidatorStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("Could not get validator status %v", err)
	}
	if resp.Status != ethpb.ValidatorStatus_EXITED {
		t.Errorf("Wanted %v, got %v", ethpb.ValidatorStatus_EXITED, resp.Status)
	}
}

func TestValidatorStatus_Exited(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	ctx := context.Background()

	pubKey := pubKey(1)
	if err := db.SaveValidatorIndex(ctx, pubKey, 0); err != nil {
		t.Fatalf("Could not save validator index: %v", err)
	}

	// Exit because only exit epoch is =< current epoch.
	slot := uint64(10000)
	epoch := helpers.SlotToEpoch(slot)
	block := blk.NewGenesisBlock([]byte{})
	if err := db.SaveBlock(ctx, block); err != nil {
		t.Fatalf("Could not save genesis block: %v", err)
	}
	genesisRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		t.Fatalf("Could not get signing root %v", err)
	}
	numDeposits := params.BeaconConfig().MinGenesisActiveValidatorCount
	beaconState, _ := testutil.DeterministicGenesisState(t, numDeposits)
	if err := db.SaveState(ctx, beaconState, genesisRoot); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveHeadBlockRoot(ctx, genesisRoot); err != nil {
		t.Fatalf("Could not save genesis state: %v", err)
	}
	state, _ := stateTrie.InitializeFromProto(&pbp2p.BeaconState{
		Slot: slot,
		Validators: []*ethpb.Validator{{
			PublicKey:         pubKey,
			WithdrawableEpoch: epoch + 1},
		},
	})
	depData := &ethpb.Deposit_Data{
		PublicKey:             pubKey,
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
	}

	deposit := &ethpb.Deposit{
		Data: depData,
	}
	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()
	depositCache.InsertDeposit(ctx, deposit, 0 /*blockNum*/, 0, depositTrie.Root())
	height := time.Unix(int64(params.BeaconConfig().Eth1FollowDistance), 0).Unix()
	p := &mockPOW.POWChain{
		TimesByHeight: map[int]uint64{
			0: uint64(height),
		},
	}
	vs := &Server{
		BeaconDB:          db,
		ChainStartFetcher: p,
		Eth1InfoFetcher:   p,
		BlockFetcher:      p,
		DepositFetcher:    depositCache,
		HeadFetcher:       &mockChain.ChainService{State: state, Root: genesisRoot[:]},
	}
	req := &ethpb.ValidatorStatusRequest{
		PublicKey: pubKey,
	}
	resp, err := vs.ValidatorStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("Could not get validator status %v", err)
	}
	if resp.Status != ethpb.ValidatorStatus_EXITED {
		t.Errorf("Wanted %v, got %v", ethpb.ValidatorStatus_EXITED, resp.Status)
	}
}

func TestValidatorStatus_UnknownStatus(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	pubKey := pubKey(1)
	depositCache := depositcache.NewDepositCache()
	stateObj, err := stateTrie.InitializeFromProtoUnsafe(&pbp2p.BeaconState{
		Slot: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	vs := &Server{
		DepositFetcher:  depositCache,
		Eth1InfoFetcher: &mockPOW.POWChain{},
		HeadFetcher: &mockChain.ChainService{
			State: stateObj,
		},
		BeaconDB: db,
	}
	req := &ethpb.ValidatorStatusRequest{
		PublicKey: pubKey,
	}
	resp, err := vs.ValidatorStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("Could not get validator status %v", err)
	}
	if resp.Status != ethpb.ValidatorStatus_UNKNOWN_STATUS {
		t.Errorf("Wanted %v, got %v", ethpb.ValidatorStatus_UNKNOWN_STATUS, resp.Status)
	}
}

func TestMultipleValidatorStatus_OK(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	ctx := context.Background()

	pubKeys := [][]byte{pubKey(1), pubKey(2), pubKey(3)}
	stateObj, err := stateTrie.InitializeFromProtoUnsafe(&pbp2p.BeaconState{
		Slot: 4000,
		Validators: []*ethpb.Validator{
			{
				ActivationEpoch: 0,
				ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
				PublicKey:       pubKeys[0],
			},
			{
				ActivationEpoch: 0,
				ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
				PublicKey:       pubKeys[1],
			},
		},
	})
	block := blk.NewGenesisBlock([]byte{})
	genesisRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		t.Fatalf("Could not get signing root %v", err)
	}
	depData := &ethpb.Deposit_Data{
		PublicKey:             pubKey(1),
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
		Amount:                10,
	}

	dep := &ethpb.Deposit{
		Data: depData,
	}
	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()
	depositCache.InsertDeposit(ctx, dep, 10 /*blockNum*/, 0, depositTrie.Root())
	depData = &ethpb.Deposit_Data{
		PublicKey:             pubKey(3),
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
		Amount:                10,
	}

	dep = &ethpb.Deposit{
		Data: depData,
	}
	depositTrie.Insert(dep.Data.Signature, 15)
	depositCache.InsertDeposit(context.Background(), dep, 0, 0, depositTrie.Root())

	if err := db.SaveValidatorIndex(ctx, pubKeys[0], 0); err != nil {
		t.Fatalf("could not save validator index: %v", err)
	}
	if err := db.SaveValidatorIndex(ctx, pubKeys[1], 1); err != nil {
		t.Fatalf("could not save validator index: %v", err)
	}
	vs := &Server{
		BeaconDB:           db,
		Ctx:                context.Background(),
		CanonicalStateChan: make(chan *pbp2p.BeaconState, 1),
		ChainStartFetcher:  &mockPOW.POWChain{},
		BlockFetcher:       &mockPOW.POWChain{},
		Eth1InfoFetcher:    &mockPOW.POWChain{},
		DepositFetcher:     depositCache,
		HeadFetcher:        &mockChain.ChainService{State: stateObj, Root: genesisRoot[:]},
	}
	activeExists, response, err := vs.multipleValidatorStatus(context.Background(), pubKeys)
	if err != nil {
		t.Fatal(err)
	}
	if !activeExists {
		t.Fatal("No activated validator exists when there was supposed to be 2")
	}
	if response[0].Status.Status != ethpb.ValidatorStatus_ACTIVE {
		t.Errorf("Validator with pubkey %#x is not activated and instead has this status: %s",
			response[0].PublicKey, response[0].Status.Status.String())
	}

	if response[1].Status.Status != ethpb.ValidatorStatus_ACTIVE {
		t.Errorf("Validator with pubkey %#x was activated when not supposed to",
			response[1].PublicKey)
	}

	if response[2].Status.Status != ethpb.ValidatorStatus_DEPOSITED {
		t.Errorf("Validator with pubkey %#x is not activated and instead has this status: %s",
			response[2].PublicKey, response[2].Status.Status.String())
	}
}

func TestValidatorStatus_CorrectActivationQueue(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	ctx := context.Background()

	pbKey := pubKey(5)
	if err := db.SaveValidatorIndex(ctx, pbKey, 5); err != nil {
		t.Fatalf("Could not save validator index: %v", err)
	}
	block := blk.NewGenesisBlock([]byte{})
	if err := db.SaveBlock(ctx, block); err != nil {
		t.Fatalf("Could not save genesis block: %v", err)
	}
	genesisRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		t.Fatalf("Could not get signing root %v", err)
	}
	currentSlot := uint64(5000)
	// Pending active because activation epoch is still defaulted at far future slot.
	state, err := stateTrie.InitializeFromProtoUnsafe(&pbp2p.BeaconState{
		Validators: []*ethpb.Validator{
			{
				ActivationEpoch:   0,
				PublicKey:         pubKey(0),
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
			{
				ActivationEpoch:   0,
				PublicKey:         pubKey(1),
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
			{
				ActivationEpoch:   0,
				PublicKey:         pubKey(2),
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
			{
				ActivationEpoch:   0,
				PublicKey:         pubKey(3),
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
			{
				ActivationEpoch:   currentSlot/params.BeaconConfig().SlotsPerEpoch + 1,
				PublicKey:         pbKey,
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
			{
				ActivationEpoch:   currentSlot/params.BeaconConfig().SlotsPerEpoch + 4,
				PublicKey:         pubKey(5),
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
		},
		Slot: currentSlot,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveState(ctx, state, genesisRoot); err != nil {
		t.Fatalf("could not save state: %v", err)
	}
	if err := db.SaveHeadBlockRoot(ctx, genesisRoot); err != nil {
		t.Fatalf("Could not save genesis state: %v", err)
	}

	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()

	for i := 0; i < 6; i++ {
		depData := &ethpb.Deposit_Data{
			PublicKey:             pubKey(uint64(i)),
			Signature:             []byte("hi"),
			WithdrawalCredentials: []byte("hey"),
		}

		deposit := &ethpb.Deposit{
			Data: depData,
		}
		depositCache.InsertDeposit(ctx, deposit, 0 /*blockNum*/, 0, depositTrie.Root())

	}

	height := time.Unix(int64(params.BeaconConfig().Eth1FollowDistance), 0).Unix()
	p := &mockPOW.POWChain{
		TimesByHeight: map[int]uint64{
			0: uint64(height),
		},
	}
	vs := &Server{
		BeaconDB:          db,
		ChainStartFetcher: p,
		BlockFetcher:      p,
		Eth1InfoFetcher:   p,
		DepositFetcher:    depositCache,
		HeadFetcher:       &mockChain.ChainService{State: state, Root: genesisRoot[:]},
	}
	req := &ethpb.ValidatorStatusRequest{
		PublicKey: pbKey,
	}
	resp, err := vs.ValidatorStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("Could not get validator status %v", err)
	}
	if resp.Status != ethpb.ValidatorStatus_PENDING {
		t.Errorf("Wanted %v, got %v", ethpb.ValidatorStatus_PENDING, resp.Status)
	}
	if resp.PositionInActivationQueue != 2 {
		t.Errorf("Expected Position in activation queue of %d but instead got %d", 2, resp.PositionInActivationQueue)
	}
}

func TestDepositBlockSlotAfterGenesisTime(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	ctx := context.Background()

	pubKey := pubKey(1)
	if err := db.SaveValidatorIndex(ctx, pubKey, 0); err != nil {
		t.Fatalf("Could not save validator index: %v", err)
	}

	depData := &ethpb.Deposit_Data{
		PublicKey:             pubKey,
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
	}

	deposit := &ethpb.Deposit{
		Data: depData,
	}
	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()
	depositCache.InsertDeposit(ctx, deposit, 0 /*blockNum*/, 0, depositTrie.Root())

	timestamp := time.Unix(int64(params.BeaconConfig().Eth1FollowDistance), 0).Unix()
	p := &mockPOW.POWChain{
		TimesByHeight: map[int]uint64{
			int(params.BeaconConfig().Eth1FollowDistance): uint64(timestamp),
		},
	}

	block := blk.NewGenesisBlock([]byte{})
	if err := db.SaveBlock(ctx, block); err != nil {
		t.Fatalf("Could not save genesis block: %v", err)
	}
	genesisRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		t.Fatalf("Could not get signing root %v", err)
	}

	activeEpoch := helpers.ActivationExitEpoch(0)

	state, err := stateTrie.InitializeFromProtoUnsafe(&pbp2p.BeaconState{
		GenesisTime: uint64(time.Unix(0, 0).Unix()),
		Slot:        10000,
		Validators: []*ethpb.Validator{{
			ActivationEpoch: activeEpoch,
			ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
			PublicKey:       pubKey},
		}})
	if err != nil {
		t.Fatal(err)
	}

	vs := &Server{
		BeaconDB:          db,
		ChainStartFetcher: p,
		BlockFetcher:      p,
		Eth1InfoFetcher:   p,
		DepositFetcher:    depositCache,
		HeadFetcher:       &mockChain.ChainService{State: state, Root: genesisRoot[:]},
	}

	eth1BlockNumBigInt := big.NewInt(1000000)

	resp, err := vs.depositBlockSlot(context.Background(), eth1BlockNumBigInt, state)
	if err != nil {
		t.Fatalf("Could not get the deposit block slot %v", err)
	}

	expected := uint64(53)

	if resp != expected {
		t.Errorf("Wanted %v, got %v", expected, resp)
	}
}

func TestDepositBlockSlotBeforeGenesisTime(t *testing.T) {
	db := dbutil.SetupDB(t)
	defer dbutil.TeardownDB(t, db)
	ctx := context.Background()

	pubKey := pubKey(1)
	if err := db.SaveValidatorIndex(ctx, pubKey, 0); err != nil {
		t.Fatalf("Could not save validator index: %v", err)
	}

	depData := &ethpb.Deposit_Data{
		PublicKey:             pubKey,
		Signature:             []byte("hi"),
		WithdrawalCredentials: []byte("hey"),
	}

	deposit := &ethpb.Deposit{
		Data: depData,
	}
	depositTrie, err := trieutil.NewTrie(int(params.BeaconConfig().DepositContractTreeDepth))
	if err != nil {
		t.Fatalf("Could not setup deposit trie: %v", err)
	}
	depositCache := depositcache.NewDepositCache()
	depositCache.InsertDeposit(ctx, deposit, 0 /*blockNum*/, 0, depositTrie.Root())

	timestamp := time.Unix(int64(params.BeaconConfig().Eth1FollowDistance), 0).Unix()
	p := &mockPOW.POWChain{
		TimesByHeight: map[int]uint64{
			int(params.BeaconConfig().Eth1FollowDistance): uint64(timestamp),
		},
	}

	block := blk.NewGenesisBlock([]byte{})
	if err := db.SaveBlock(ctx, block); err != nil {
		t.Fatalf("Could not save genesis block: %v", err)
	}
	genesisRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		t.Fatalf("Could not get signing root %v", err)
	}

	activeEpoch := helpers.ActivationExitEpoch(0)

	state, err := stateTrie.InitializeFromProtoUnsafe(&pbp2p.BeaconState{
		GenesisTime: uint64(time.Unix(25000, 0).Unix()),
		Slot:        10000,
		Validators: []*ethpb.Validator{{
			ActivationEpoch: activeEpoch,
			ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
			PublicKey:       pubKey},
		}})
	if err != nil {
		t.Fatal(err)
	}

	vs := &Server{
		BeaconDB:          db,
		ChainStartFetcher: p,
		BlockFetcher:      p,
		Eth1InfoFetcher:   p,
		DepositFetcher:    depositCache,
		HeadFetcher:       &mockChain.ChainService{State: state, Root: genesisRoot[:]},
	}

	eth1BlockNumBigInt := big.NewInt(1000000)

	resp, err := vs.depositBlockSlot(context.Background(), eth1BlockNumBigInt, state)
	if err != nil {
		t.Fatalf("Could not get the deposit block slot %v", err)
	}

	expected := uint64(0)

	if resp != expected {
		t.Errorf("Wanted %v, got %v", expected, resp)
	}
}
