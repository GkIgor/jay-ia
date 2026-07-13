package memory_test

import (
	"errors"
	"testing"

	"github.com/GkIgor/jay-ia/core/internal/memory"
)

func TestInMemoryStore_GetPutDelete(t *testing.T) {
	store := memory.NewInMemoryStore()

	// 1. Get non-existent key -> ErrNotFound
	val, err := store.Get("nonexistent")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Fatalf("expected memory.ErrNotFound, got %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}

	// 2. Put key -> no error
	err = store.Put("key1", "value1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 3. Get key -> success
	val, err = store.Get("key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "value1" {
		t.Fatalf("expected value1, got %v", val)
	}

	// 4. Update key -> success
	err = store.Put("key1", "value2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	val, err = store.Get("key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "value2" {
		t.Fatalf("expected value2, got %v", val)
	}

	// 5. Delete key -> success
	err = store.Delete("key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 6. Get deleted key -> ErrNotFound
	val, err = store.Get("key1")
	if !errors.Is(err, memory.ErrNotFound) {
		t.Fatalf("expected memory.ErrNotFound, got %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}

	// 7. Delete non-existent key -> no error
	err = store.Delete("key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMemoryStoreInterface(t *testing.T) {
	// Verify that InMemoryStore implements MemoryStore interface
	var _ memory.MemoryStore = memory.NewInMemoryStore()
}
