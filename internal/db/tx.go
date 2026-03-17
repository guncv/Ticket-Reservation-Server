package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PgTx struct {
	tx pgx.Tx
}

func (t *PgTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return t.tx.Begin(ctx)
}

func (t *PgTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *PgTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *PgTx) Exec(
	ctx context.Context,
	sql string,
	arguments ...any,
) (commandTag pgconn.CommandTag, err error) {
	return t.tx.Exec(ctx, sql, arguments...)
}

func (t *PgTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.tx.Query(ctx, sql, args...)
}

func (t *PgTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}

func (t *PgTx) Release() {}
