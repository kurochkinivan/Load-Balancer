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
	log           *slog.Logger
	clientStorage ClientStorage
}

func New(log *slog.Logger, clientStorage ClientStorage) *ClientsUseCase {
	return &ClientsUseCase{
		log:           log,
		clientStorage: clientStorage,
	}
}

type ClientStorage interface {
	GetClients(ctx context.Context) ([]*entity.Client, error)
	CreateClient(ctx context.Context, client *entity.Client) error
	DeleteClient(ctx context.Context, ipAdress string) error
}

func (c *ClientsUseCase) GetClients(ctx context.Context) ([]*entity.Client, error) {
	const op = "ClientsUseCase.GetClients"

	clients, err := c.clientStorage.GetClients(ctx)
	if err != nil {
		c.log.Error("failed to get clients", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return clients, nil
}

func (c *ClientsUseCase) CreateClient(ctx context.Context, client *entity.Client) error {
	const op = "ClientsUseCase.CreateClient"

	err := c.clientStorage.CreateClient(ctx, client)
	if err != nil {
		if errors.Is(err, storage.ErrClientExists) {
			c.log.Error("client already exists")
			return fmt.Errorf("%s: %w", op, ErrClientExists)
		}

		c.log.Error("failed to create client", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *ClientsUseCase) DeleteClient(ctx context.Context, ipAdress string) error {
	const op = "ClientsUseCase.DeleteClient"

	err := c.clientStorage.DeleteClient(ctx, ipAdress)
	if err != nil {
		if errors.Is(err, storage.ErrClientNotFound) {
			c.log.Warn("client was not found")
			return fmt.Errorf("%s: %w", op, ErrClientNotFound)
		}

		c.log.Error("failed to delete client", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
