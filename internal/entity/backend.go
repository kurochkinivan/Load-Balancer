package entity

import (
	"net/url"
	"sync/atomic"
)

// Backend represents a backend server.
type Backend struct {
	URL       *url.URL
	available atomic.Bool
}

// NewBackend creates a new Backend instance.
func NewBackend(URL *url.URL) *Backend {
	b := &Backend{
		URL:       URL,
		available: atomic.Bool{},
	}
	b.available.Store(true)

	return b
}

// SetAvailable sets the availability of the backend.
//
// This method is concurrently safe.
func (b *Backend) SetAvailable(available bool) {
	b.available.Store(available)
}

// IsAvailable returns true if the backend is available, false otherwise.
//
// This method is concurrently safe.
func (b *Backend) IsAvailable() bool {
	return b.available.Load()
}
