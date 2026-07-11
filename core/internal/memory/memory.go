package memory

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("key not found")

// MemoryStore defines the interface for the minimal memory system
type MemoryStore interface {
	Get(key string) (any, error)
	Put(key string, value any) error
	Delete(key string) error
}

// InMemoryStore is an ephemeral implementation of MemoryStore
type InMemoryStore struct {
	mu    sync.RWMutex
	store map[string]any
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		store: make(map[string]any),
	}
}

func (s *InMemoryStore) Get(key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.store[key]
	if !ok {
		return nil, ErrNotFound
	}
	return val, nil
}

func (s *InMemoryStore) Put(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = value
	return nil
}

func (s *InMemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, key)
	return nil
}
