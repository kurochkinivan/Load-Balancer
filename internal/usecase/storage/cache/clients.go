package cache

import (
	"container/list"
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/kurochkinivan/load_balancer/internal/entity"
)

type LRUClientCache struct {
	log         *slog.Logger
	maxElements int
	mu          *sync.Mutex
	list        *list.List                // least frequently used - the back of the list
	items       map[string]*list.Element  // string (ipAdress) -> element in list
	cache       map[string]*entity.Client // string (ipAdress) -> *entity.Client
}

func NewClientsCache(log *slog.Logger, maxElements int) *LRUClientCache {
	if maxElements < 0 {
		maxElements = 0
	}

	return &LRUClientCache{
		log:         log,
		maxElements: maxElements,
		mu:          new(sync.Mutex),
		list:        list.New(),
		items:       make(map[string]*list.Element),
		cache:       make(map[string]*entity.Client),
	}
}

func (c *LRUClientCache) Client(ipAddress string) (*entity.Client, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[ipAddress]
	if ok {
		c.list.MoveToFront(el)
		return c.cache[ipAddress], true
	}

	return nil, false
}

func (c *LRUClientCache) AddClient(client *entity.Client) {
	if c.maxElements == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[client.IPAddress]; ok {
		c.cache[client.IPAddress] = client
		c.list.MoveToFront(el)
		return
	}

	if len(c.items) >= c.maxElements {
		ipAddress := c.list.Remove(c.list.Back()).(string)
		delete(c.items, ipAddress)
		delete(c.cache, ipAddress)

	}

	c.cache[client.IPAddress] = client
	el := c.list.PushFront(client.IPAddress)
	c.items[client.IPAddress] = el
}

func (c *LRUClientCache) DeleteClient(ipAddress string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[ipAddress]; ok {
		c.list.Remove(el)
		delete(c.items, ipAddress)
		delete(c.cache, ipAddress)
	}
}

func (c *LRUClientCache) refillAllClients() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, client := range c.cache {
		client.RefillTokensOncePerSecond()
	}
}

func (c *LRUClientCache) StartTokenRefiller(ctx context.Context) {
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
