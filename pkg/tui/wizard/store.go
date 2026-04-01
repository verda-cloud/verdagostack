package wizard

import (
	"maps"
	"sync"
)

// Store is the engine's shared data layer. Regions read from it,
// the engine and loaders write to it.
type Store struct {
	collected map[string]any
	data      map[string]any
	mu        sync.RWMutex
}

// NewStore creates an empty store.
func NewStore() *Store {
	return &Store{
		collected: make(map[string]any),
		data:      make(map[string]any),
	}
}

// Collected returns a snapshot of the wizard step values.
func (s *Store) Collected() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap := make(map[string]any, len(s.collected))
	maps.Copy(snap, s.collected)
	return snap
}

// SetCollected sets a wizard step value.
func (s *Store) SetCollected(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collected[key] = value
}

// ClearCollected removes a wizard step value.
func (s *Store) ClearCollected(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.collected, key)
}

// Get reads an arbitrary value from the store.
func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// Set writes an arbitrary value to the store.
func (s *Store) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}
