package entity

import (
	"net/url"
	"sync"
)

// Backend represents a backend server.
type Backend struct {
	URL       *url.URL
	mu        *sync.RWMutex
	available bool
}

// NewBackend creates a new Backend instance.
func NewBackend(URL *url.URL) *Backend {
	return &Backend{
		URL:       URL,
		mu:        new(sync.RWMutex),
		available: true,
	}
}

// SetAvailable sets the availability of the backend.
//
// This method is concurrently safe.
func (b *Backend) SetAvailable(available bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.available = available
}

// IsAvailable returns true if the backend is available, false otherwise.
//
// This method is concurrently safe.
func (b *Backend) IsAvailable() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.available
}
