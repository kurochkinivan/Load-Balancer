package pg

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TableClients = "clients"
)

type Storage struct {
	pool *pgxpool.Pool
	qb   sq.StatementBuilderType
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{
		pool: pool,
		qb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}
