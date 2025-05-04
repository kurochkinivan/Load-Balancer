package pg

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/kurochkinivan/load_balancer/internal/usecase/storage"
	"github.com/kurochkinivan/load_balancer/pkg/pgerr"
)

func (s *Storage) Client(ctx context.Context, ipAdress string) (*entity.Client, error) {
	const op = "storage.pg.Client"

	sql, args, err := s.qb.
		Select("id",
			"ip_address",
			"name",
			"capacity",
			"rate_per_second").
		From(TableClients).
		Where(sq.Eq{
			"ip_address": ipAdress,
		}).
		ToSql()
	if err != nil {
		return nil, pgerr.ErrCreateQuery(op, err)
	}

	row := s.pool.QueryRow(ctx, sql, args...)

	client := new(entity.Client)
	err = row.Scan(
		&client.ID,
		&client.IPAddress,
		&client.Name,
		&client.Capacity,
		&client.RatePerSecond,
	)
	client.Tokens.Store(client.Capacity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrClientNotFound
		}

		return nil, pgerr.ErrScan(op, err)
	}

	return client, nil
}

func (s *Storage) Clients(ctx context.Context) ([]*entity.Client, error) {
	const op = "storage.pg.Clients"

	sql, args, err := s.qb.
		Select("id",
			"ip_address",
			"name",
			"capacity",
			"rate_per_second").
		From(TableClients).
		ToSql()
	if err != nil {
		return nil, pgerr.ErrCreateQuery(op, err)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgerr.ErrDoQuery(op, err)
	}
	defer rows.Close()

	clients := make([]*entity.Client, 0)

	for rows.Next() {
		client := new(entity.Client)
		err = rows.Scan(
			&client.ID,
			&client.IPAddress,
			&client.Name,
			&client.Capacity,
			&client.RatePerSecond,
		)
		if err != nil {
			return nil, pgerr.ErrScan(op, err)
		}
		client.Tokens.Store(client.Capacity)
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, pgerr.ErrScan(op, err)
	}

	return clients, nil
}

func (s *Storage) CreateClient(ctx context.Context, client *entity.Client) error {
	const op = "storage.pg.CreateClient"

	sql, args, err := s.qb.
		Insert(TableClients).
		Columns(
			"ip_address",
			"name",
			"capacity",
			"rate_per_second",
		).
		Values(
			client.IPAddress,
			client.Name,
			client.Capacity,
			client.RatePerSecond,
		).
		Suffix("ON CONFLICT (ip_address) DO NOTHING").
		ToSql()
	if err != nil {
		return pgerr.ErrCreateQuery(op, err)
	}

	cmdTag, err := s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return pgerr.ErrExec(op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return storage.ErrClientExists
	}

	return nil
}

func (s *Storage) DeleteClient(ctx context.Context, ipAddress string) error {
	const op = "storage.pg.DeleteClient"

	sql, args, err := s.qb.
		Delete(TableClients).
		Where(sq.Eq{"ip_address": ipAddress}).
		ToSql()
	if err != nil {
		return pgerr.ErrCreateQuery(op, err)
	}

	cmd, err := s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return pgerr.ErrExec(op, err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrClientNotFound)
	}

	return nil
}
