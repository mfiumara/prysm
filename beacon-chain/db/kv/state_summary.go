package kv

import (
	"context"

	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

// SaveStateSummary saves a state summary object to the DB.
func (k *Store) SaveStateSummary(ctx context.Context, summary *pb.StateSummary) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.SaveStateSummary")
	defer span.End()

	enc, err := encode(summary)
	if err != nil {
		return err
	}
	return k.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(stateSummaryBucket)
		return bucket.Put(summary.Root, enc)
	})
}

// SaveStateSummaries saves state summary objects to the DB.
func (k *Store) SaveStateSummaries(ctx context.Context, summaries []*pb.StateSummary) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.SaveStateSummaries")
	defer span.End()

	return k.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(stateSummaryBucket)
		for _, summary := range summaries {
			enc, err := encode(summary)
			if err != nil {
				return err
			}
			if err := bucket.Put(summary.Root, enc); err != nil {
				return err
			}
		}
		return nil
	})
}

// StateSummary returns the state summary object from the db using input block root.
func (k *Store) StateSummary(ctx context.Context, blockRoot [32]byte) (*pb.StateSummary, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.StateSummary")
	defer span.End()

	var summary *pb.StateSummary
	err := k.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(stateSummaryBucket)
		enc := bucket.Get(blockRoot[:])
		if enc == nil {
			return nil
		}
		summary = &pb.StateSummary{}
		return decode(enc, summary)
	})

	return summary, err
}

// HasStateSummary returns true if a state summary exists in DB.
func (k *Store) HasStateSummary(ctx context.Context, blockRoot [32]byte) bool {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.HasStateSummary")
	defer span.End()
	var exists bool
	// #nosec G104. Always returns nil.
	k.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(stateSummaryBucket)
		exists = bucket.Get(blockRoot[:]) != nil
		return nil
	})
	return exists
}
