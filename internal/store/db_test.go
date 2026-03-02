package store_test

import (
	"testing"

	"github.com/baochen10luo/stagenthand/internal/store"
)

func TestNew_CreatesTablesOK(t *testing.T) {
	db, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("store.New() error = %v", err)
	}
	if db == nil {
		t.Fatal("store.New() returned nil DB")
	}
}
