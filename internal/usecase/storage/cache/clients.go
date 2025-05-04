package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/kurochkinivan/load_balancer/internal/entity"
)

type ClientsCache struct {
	log     *slog.Logger
	clients *sync.Map // string (ip_adress) -> *entity.Client
}

func NewClientsCache(log *slog.Logger) *ClientsCache {
	return &ClientsCache{
		log:     log,
		clients: new(sync.Map),
	}
}

func (c *ClientsCache) Client(ip_address string) (*entity.Client, bool) {
	val, ok := c.clients.Load(ip_address)
	if !ok {
		return nil, false
	}
	return val.(*entity.Client), true
}

func (c *ClientsCache) AddClient(client *entity.Client) {
	c.clients.Store(client.IPAddress, client)
}

func (c *ClientsCache) DeleteClient(ip_address string) {
	c.clients.Delete(ip_address)
}

func (c *ClientsCache) refillAllClients() {
	c.clients.Range(func(key, value any) bool {
		client := value.(*entity.Client)
		client.RefillTokensOncePerSecond()
		return true
	})
}

func (c *ClientsCache) StartTokenRefiller(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	c.log.Info("starting token refiller...")

	for {
		select {
		case <-ticker.C:
			c.refillAllClients()
		case <-ctx.Done():
			c.log.Info("token refiller is terminated due to context cancellation")
			return
		}
	}
}
