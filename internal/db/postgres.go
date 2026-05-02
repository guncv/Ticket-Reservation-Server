package db

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolStats contains connection pool statistics for monitoring
type PoolStats struct {
	TotalConns        int32         `json:"total_conns"`         // Total number of connections in the pool
	AcquiredConns     int32         `json:"acquired_conns"`      // Connections currently in use
	IdleConns         int32         `json:"idle_conns"`          // Connections currently idle
	MaxConns          int32         `json:"max_conns"`           // Maximum pool size
	AcquireCount      int64         `json:"acquire_count"`       // Total successful acquires
	AcquireDuration   time.Duration `json:"acquire_duration_ns"` // Total time spent acquiring
	EmptyAcquireCount int64         `json:"empty_acquire_count"` // Acquires that had to wait (pool was empty)
	CanceledAcquires  int64         `json:"canceled_acquires"`   // Acquires canceled by context
	ConstructingConns int32         `json:"constructing_conns"`  // Connections being created
	PoolUtilization   float64       `json:"pool_utilization"`    // Percentage of pool in use (0-1)
}

type Tx interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	Exec(
		ctx context.Context,
		sql string,
		arguments ...any,
	) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Release()
}

type Conn interface {
	Exec(
		ctx context.Context,
		sql string,
		arguments ...any,
	) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Release()
}

type PgPool struct {
	*pgxpool.Pool
}

func connStr(cfg config.DatabaseConfig) (string, error) {
	if cfg.ApplicationName == "" {
		return "", fmt.Errorf("cannot create postgres connection string: application name is empty")
	}
	if cfg.Host == "" {
		return "", fmt.Errorf("cannot create postgres connection string: host is empty")
	}
	if cfg.Port == "" {
		return "", fmt.Errorf("cannot create postgres connection string: port is empty")
	}
	if cfg.DbName == "" {
		return "", fmt.Errorf("cannot create postgres connection string: database name is empty")
	}
	if cfg.User == "" {
		return "", fmt.Errorf("cannot create postgres connection string: user is empty")
	}
	if cfg.Password == "" {
		return "", fmt.Errorf("cannot create postgres connection string: password is empty")
	}
	if cfg.SSLMode == "" {
		return "", fmt.Errorf("cannot create postgres connection string: ssl")
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.DbName,
	}

	q := u.Query()
	q.Set("application_name", cfg.ApplicationName)
	q.Set("sslmode", cfg.SSLMode)
	u.RawQuery = q.Encode()

	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = time.Second * 5
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 50000
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 5000
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = time.Hour
	}
	if cfg.ConnMaxLifetimeJitter == 0 {
		cfg.ConnMaxLifetimeJitter = cfg.ConnMaxLifetime / 8
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = time.Minute * 15
	}
	if cfg.HealthCheckPeriod == 0 {
		cfg.HealthCheckPeriod = time.Minute
	}

	if cfg.EventTimeout == 0 {
		cfg.EventTimeout = time.Second * 5
	}

	return u.String(), nil
}

func NewPgPool(cfg *config.Config) (*PgPool, error) {
	dbCfg := cfg.DatabaseConfig

	connStr, err := connStr(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection string: %w", err)
	}
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	poolConfig.MaxConns = int32(dbCfg.MaxOpenConns)
	poolConfig.MinConns = int32(dbCfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = dbCfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = dbCfg.ConnMaxIdleTime
	poolConfig.HealthCheckPeriod = dbCfg.HealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = dbCfg.ConnectTimeout

	ctx, cancel := context.WithTimeout(context.Background(), dbCfg.HealthCheckPeriod)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &PgPool{
		Pool: pool,
	}, nil
}

type dbKey struct{}

func (c *PgPool) EnsureTxFromCtx(ctx context.Context) (context.Context, Tx, error) {
	txValue := ctx.Value(dbKey{})
	if txValue == nil {
		tx, err := c.Pool.Begin(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
		}
		pgTx := &PgTx{tx: tx}
		return context.WithValue(ctx, dbKey{}, pgTx), pgTx, nil
	}

	tx, ok := txValue.(*PgTx)
	if !ok {
		conn, ok := txValue.(*PgConn)
		if !ok {
			return nil, nil, fmt.Errorf("tx is not a valid transaction")
		}

		tx, err := conn.conn.Begin(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to begin nested transaction: %w", err)
		}
		return context.WithValue(ctx, dbKey{}, &PgTx{tx: tx}), &PgTx{tx: tx}, nil
	}

	nestedTx, err := tx.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin nested transaction: %w", err)
	}

	pgTx := &PgTx{tx: nestedTx}
	return context.WithValue(ctx, dbKey{}, pgTx), pgTx, nil
}

func (c *PgPool) EnsureConnFromCtx(ctx context.Context) (context.Context, Conn, error) {
	txValue := ctx.Value(dbKey{})
	if txValue == nil {
		conn, err := c.Pool.Acquire(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to acquire connection: %w", err)
		}
		return context.WithValue(ctx, dbKey{}, &PgConn{conn: conn}), &PgConn{conn: conn}, nil
	}

	conn, ok := txValue.(Conn)
	if !ok {
		return nil, nil, fmt.Errorf("conn is not a valid connection")
	}

	return ctx, conn, nil
}

// Stats returns current connection pool statistics
func (c *PgPool) Stats() PoolStats {
	s := c.Pool.Stat()

	var utilization float64
	if s.MaxConns() > 0 {
		utilization = float64(s.AcquiredConns()) / float64(s.MaxConns())
	}

	return PoolStats{
		TotalConns:        s.TotalConns(),
		AcquiredConns:     s.AcquiredConns(),
		IdleConns:         s.IdleConns(),
		MaxConns:          s.MaxConns(),
		AcquireCount:      s.AcquireCount(),
		AcquireDuration:   s.AcquireDuration(),
		EmptyAcquireCount: s.EmptyAcquireCount(),
		CanceledAcquires:  s.CanceledAcquireCount(),
		ConstructingConns: s.ConstructingConns(),
		PoolUtilization:   utilization,
	}
}
