package testing

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path"
	"testing"

	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/slasher/db"
)

// SetupSlasherDB instantiates and returns a SlasherDB instance.
func SetupSlasherDB(t testing.TB) *db.Store {

	randPath, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		t.Fatalf("Could not generate random file path: %v", err)
	}
	p := path.Join(testutil.TempDir(), fmt.Sprintf("/%d", randPath))
	if err := os.RemoveAll(p); err != nil {
		t.Fatalf("Failed to remove directory: %v", err)
	}
	db, err := db.NewDB(p)

	if err != nil {
		t.Fatalf("Failed to instantiate DB: %v", err)
	}
	return db
}

// TeardownSlasherDB cleans up a test BeaconDB instance.
func TeardownSlasherDB(t testing.TB, db *db.Store) {
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}
	if err := os.RemoveAll(db.DatabasePath()); err != nil {
		t.Fatalf("Failed to remove directory: %v", err)
	}
}
