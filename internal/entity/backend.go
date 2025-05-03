package entity

import (
	"net/url"
	"sync"
)

type Backend struct {
	URL       *url.URL
	mu        *sync.RWMutex
	available bool
}

func NewBackend(URL *url.URL) *Backend {
	return &Backend{
		URL:       URL,
		mu:        new(sync.RWMutex),
		available: true,
	}
}

func (b *Backend) SetAvailable(available bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.available = available
}

func (b *Backend) IsAvailable() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.available
}
