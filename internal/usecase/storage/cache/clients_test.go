package cache

import (
	"log/slog"
	"os"
	"sync/atomic"
	"testing"

	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewClientsCache(t *testing.T) {
	tests := []struct {
		name        string
		maxElements int
		expected    int
	}{
		{
			name:        "Normal initialization",
			maxElements: 10,
			expected:    10,
		},
		{
			name:        "Zero size",
			maxElements: 0,
			expected:    0,
		},
		{
			name:        "Negative size converted to zero",
			maxElements: -5,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			cache := NewClientsCache(logger, tt.maxElements)

			assert.NotNil(t, cache)
			assert.Equal(t, tt.expected, cache.maxElements)
			assert.NotNil(t, cache.mu)
			assert.NotNil(t, cache.list)
			assert.NotNil(t, cache.items)
			assert.NotNil(t, cache.cache)
			assert.Empty(t, cache.items)
			assert.Empty(t, cache.cache)
		})
	}
}

func TestLRUClientCache_Client(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cache := NewClientsCache(logger, 2)

	// Create test clients
	client1 := &entity.Client{IPAddress: "192.168.1.1"}
	client2 := &entity.Client{IPAddress: "192.168.1.2"}

	// Add clients to cache
	cache.UpdateClient(client1)
	cache.UpdateClient(client2)

	tests := []struct {
		name      string
		ipAddress string
		wantFound bool
	}{
		{
			name:      "Get existing client",
			ipAddress: "192.168.1.1",
			wantFound: true,
		},
		{
			name:      "Get another existing client",
			ipAddress: "192.168.1.2",
			wantFound: true,
		},
		{
			name:      "Get non-existing client",
			ipAddress: "192.168.1.3",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, found := cache.Client(tt.ipAddress)

			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.NotNil(t, client)
				assert.Equal(t, tt.ipAddress, client.IPAddress)
			} else {
				assert.Nil(t, client)
			}
		})
	}

	// Test that accessing a client moves it to the front of the list
	// First, let's get the list back to a known state
	cache = NewClientsCache(logger, 2)
	cache.UpdateClient(client1) // Most recently used
	cache.UpdateClient(client2) // Least recently used

	// Now get client1, which should move it to the front
	_, _ = cache.Client(client1.IPAddress)

	// Add a third client, which should evict client2 (least recently used)
	client3 := &entity.Client{IPAddress: "192.168.1.3"}
	cache.UpdateClient(client3)

	// Check that client2 was evicted
	_, found := cache.Client(client2.IPAddress)
	assert.False(t, found)

	// Check that client1 and client3 are still in the cache
	_, found = cache.Client(client1.IPAddress)
	assert.True(t, found)

	_, found = cache.Client(client3.IPAddress)
	assert.True(t, found)
}

func TestLRUClientCache_UpdateClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("Add clients without exceeding capacity", func(t *testing.T) {
		cache := NewClientsCache(logger, 3)

		clients := []*entity.Client{
			{IPAddress: "192.168.1.1"},
			{IPAddress: "192.168.1.2"},
			{IPAddress: "192.168.1.3"},
		}

		for _, client := range clients {
			cache.UpdateClient(client)
		}

		// Verify all clients are in the cache
		for _, client := range clients {
			cachedClient, found := cache.Client(client.IPAddress)
			assert.True(t, found)
			assert.Equal(t, client.IPAddress, cachedClient.IPAddress)
		}

		// Verify the size of the cache
		assert.Equal(t, 3, len(cache.cache))
		assert.Equal(t, 3, len(cache.items))
		assert.Equal(t, 3, cache.list.Len())
	})

	t.Run("Add clients exceeding capacity", func(t *testing.T) {
		cache := NewClientsCache(logger, 2)

		client1 := &entity.Client{IPAddress: "192.168.1.1"}
		client2 := &entity.Client{IPAddress: "192.168.1.2"}
		client3 := &entity.Client{IPAddress: "192.168.1.3"}

		// Add first two clients
		cache.UpdateClient(client1)
		cache.UpdateClient(client2)

		// Verify the cache state
		assert.Equal(t, 2, len(cache.cache))

		// Add third client, which should evict the least recently used (client1)
		cache.UpdateClient(client3)

		// Check if client1 was evicted
		_, found := cache.Client(client1.IPAddress)
		assert.False(t, found)

		// Check if client2 and client3 are still in the cache
		_, found = cache.Client(client2.IPAddress)
		assert.True(t, found)

		_, found = cache.Client(client3.IPAddress)
		assert.True(t, found)

		// Verify the cache size remains the same
		assert.Equal(t, 2, len(cache.cache))
		assert.Equal(t, 2, len(cache.items))
		assert.Equal(t, 2, cache.list.Len())
	})

	t.Run("Add client with existing IP address", func(t *testing.T) {
		cache := NewClientsCache(logger, 2)

		client1 := &entity.Client{IPAddress: "192.168.1.1"}
		client2 := &entity.Client{IPAddress: "192.168.1.1", Tokens: atomic.Int32{}} // Same IP address as client1 but different properties
		client2.Tokens.Store(10)

		cache.UpdateClient(client1)
		cache.UpdateClient(client2)

		// Verify the cache only contains one client (the updated one)
		assert.Equal(t, 1, len(cache.cache))
		assert.Equal(t, 1, len(cache.items))
		assert.Equal(t, 1, cache.list.Len())

		// Verify the client in the cache is client2
		cachedClient, found := cache.Client(client1.IPAddress)
		assert.True(t, found)
		assert.Equal(t, client2.Tokens.Load(), cachedClient.Tokens.Load())
	})

	t.Run("Add client to cache with zero capacity", func(t *testing.T) {
		cache := NewClientsCache(logger, 0)

		client := &entity.Client{IPAddress: "192.168.1.1"}
		cache.UpdateClient(client)

		// Verify the client was not added
		_, found := cache.Client(client.IPAddress)
		assert.False(t, found)

		// Verify the cache is empty
		assert.Equal(t, 0, len(cache.cache))
		assert.Equal(t, 0, len(cache.items))
		assert.Equal(t, 0, cache.list.Len())
	})
}

func TestLRUClientCache_DeleteClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cache := NewClientsCache(logger, 3)
	
	// Create and add test clients
	clients := []*entity.Client{
		{IPAddress: "192.168.1.1"},
		{IPAddress: "192.168.1.2"},
		{IPAddress: "192.168.1.3"},
	}
	
	for _, client := range clients {
		cache.UpdateClient(client)
	}
	
	t.Run("Delete existing client", func(t *testing.T) {
		// Delete client2
		cache.DeleteClient("192.168.1.2")
		
		// Verify client2 was deleted
		_, found := cache.Client("192.168.1.2")
		assert.False(t, found)
		
		// Verify the other clients are still there
		_, found = cache.Client("192.168.1.1")
		assert.True(t, found)
		
		_, found = cache.Client("192.168.1.3")
		assert.True(t, found)
		
		// Verify the cache size
		assert.Equal(t, 2, len(cache.cache))
		assert.Equal(t, 2, len(cache.items))
		assert.Equal(t, 2, cache.list.Len())
	})
	
	t.Run("Delete non-existing client", func(t *testing.T) {
		// This should not panic
		cache.DeleteClient("192.168.1.4")
		
		// Verify the cache size remains the same
		assert.Equal(t, 2, len(cache.cache))
		assert.Equal(t, 2, len(cache.items))
		assert.Equal(t, 2, cache.list.Len())
	})
	
	t.Run("Delete last client", func(t *testing.T) {
		// Delete client1 and client3
		cache.DeleteClient("192.168.1.1")
		cache.DeleteClient("192.168.1.3")
		
		// Verify the cache is empty
		assert.Equal(t, 0, len(cache.cache))
		assert.Equal(t, 0, len(cache.items))
		assert.Equal(t, 0, cache.list.Len())
	})
}