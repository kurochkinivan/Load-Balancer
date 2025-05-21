package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/kurochkinivan/load_balancer/internal/lib/sl"
	"github.com/kurochkinivan/load_balancer/internal/usecase/storage"
)

type ClientsUseCase struct {
	log     *slog.Logger
	storage ClientStorage
	cache   ClientCache
}

func New(log *slog.Logger, clientStorage ClientStorage, cache ClientCache) *ClientsUseCase {
	return &ClientsUseCase{
		log:     log,
		storage: clientStorage,
		cache:   cache,
	}
}

type ClientStorage interface {
	Clients(ctx context.Context) ([]*entity.Client, error)
	Client(ctx context.Context, ipAdress string) (*entity.Client, error)
	CreateClient(ctx context.Context, client *entity.Client) error
	DeleteClient(ctx context.Context, ipAdress string) error
}

type ClientCache interface {
	Client(ip_address string) (*entity.Client, bool)
	AddClient(client *entity.Client)
	DeleteClient(ip_address string)
}

func (c *ClientsUseCase) Client(ctx context.Context, ipAdress string) (*entity.Client, bool) {
	client, ok := c.cache.Client(ipAdress)
	if ok {
		c.log.Info("cache hit!", slog.String("ipAdress", ipAdress))
		return client, true
	}

	c.log.Info("cache miss, going to db...", slog.String("ipAdress", ipAdress))

	client, err := c.storage.Client(ctx, ipAdress)
	if err != nil {
		return nil, false
	}

	c.cache.AddClient(client)

	return client, true
}

func (c *ClientsUseCase) Clients(ctx context.Context) ([]*entity.Client, error) {
	const op = "ClientsUseCase.GetClients"

	clients, err := c.storage.Clients(ctx)
	if err != nil {
		c.log.Error("failed to get clients", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return clients, nil
}

func (c *ClientsUseCase) CreateClient(ctx context.Context, client *entity.Client) error {
	const op = "ClientsUseCase.CreateClient"

	err := c.storage.CreateClient(ctx, client)
	if err != nil {
		if errors.Is(err, storage.ErrClientExists) {
			c.log.Error("client already exists")
			return fmt.Errorf("%s: %w", op, ErrClientExists)
		}

		c.log.Error("failed to create client", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	c.cache.AddClient(client)

	return nil
}

func (c *ClientsUseCase) DeleteClient(ctx context.Context, ipAdress string) error {
	const op = "ClientsUseCase.DeleteClient"

	err := c.storage.DeleteClient(ctx, ipAdress)
	if err != nil {
		if errors.Is(err, storage.ErrClientNotFound) {
			c.log.Warn("client was not found")
			return fmt.Errorf("%s: %w", op, ErrClientNotFound)
		}

		c.log.Error("failed to delete client", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	c.cache.DeleteClient(ipAdress)

	return nil
}
