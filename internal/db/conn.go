package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgConn struct {
	conn *pgxpool.Conn
}

func (c *PgConn) Exec(
	ctx context.Context,
	sql string,
	arguments ...any,
) (commandTag pgconn.CommandTag, err error) {
	return c.conn.Exec(ctx, sql, arguments...)
}

func (c *PgConn) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return c.conn.Query(ctx, sql, args...)
}

func (c *PgConn) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return c.conn.QueryRow(ctx, sql, args...)
}

func (c *PgConn) Release() {
	c.conn.Release()
}
