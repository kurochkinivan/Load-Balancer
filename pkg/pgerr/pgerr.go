package pgerr

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
)

var (
	ErrNoRowsAffected = errors.New("no rows affected")
	ErrNoRows         = errors.New("no rows in the result set")
)

func ErrExec(op string, err error) error {
	return errors.Wrap(ParsePgErr(err), fmt.Sprint(op, ": failed to execute query"))
}

func ErrCreateQuery(op string, err error) error {
	return errors.Wrap(ParsePgErr(err), fmt.Sprint(op, ": failed to create sql query"))
}

func ErrDoQuery(op string, err error) error {
	return errors.Wrap(ParsePgErr(err), fmt.Sprint(op, ": failed to do sql query"))
}

func ErrScan(op string, err error) error {
	return errors.Wrap(ParsePgErr(err), fmt.Sprint(op, ": failed to scan from sql query result"))
}

func ErrCreateTx(op string, err error) error {
	return errors.Wrap(ParsePgErr(err), fmt.Sprint(op, ": failed to create transaction"))
}

func ErrInsertMultipleRows(op string, err error) error {
	return errors.Wrap(ParsePgErr(err), fmt.Sprint(op, ": failed to insert multiple rows"))
}

func ErrCommit(op string, err error) error {
	return errors.Wrap(ParsePgErr(err), fmt.Sprint(op, ": failed to commit transaction"))
}

func ParsePgErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return fmt.Errorf(
			"database error. message:%s, detail:%s, where:%s, sqlstate:%s",
			pgErr.Message,
			pgErr.Detail,
			pgErr.Where,
			pgErr.SQLState(),
		)
	}
	return err
}
