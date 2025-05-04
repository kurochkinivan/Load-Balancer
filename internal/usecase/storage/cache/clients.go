package cache

import (
	"sync"

	"github.com/kurochkinivan/load_balancer/internal/entity"
)

type Cache struct {
	clients *sync.Map // string (ip_adress) -> *entity.Client
}

func NewCache() *Cache {
	return &Cache{
		clients: new(sync.Map),
	}
}

func (c *Cache) Client(ip_address string) (*entity.Client, bool) {
	val, ok := c.clients.Load(ip_address)
	if !ok {
		return nil, false
	}
	return val.(*entity.Client), true
}

func (c *Cache) AddClient(client *entity.Client) {
	c.clients.Store(client.IPAddress, client)
}

func (c *Cache) DeleteClient(ip_address string) {
	c.clients.Delete(ip_address)
}
