package pg

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/kurochkinivan/load_balancer/internal/usecase/storage"
	"github.com/kurochkinivan/load_balancer/pkg/pgerr"
)

func (s *Storage) GetClients(ctx context.Context) ([]*entity.Client, error) {
	const op = "storage.pg.GetClients"

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
