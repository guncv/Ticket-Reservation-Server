package db

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Seed struct {
	Name      string
	Up        func(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error
	Down      func(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error
	Timestamp time.Time
}

type SeedRegistry struct {
	pool *pgxpool.Pool
}

func NewSeedRegistry(pool *pgxpool.Pool) *SeedRegistry {
	return &SeedRegistry{pool: pool}
}

func (sr *SeedRegistry) IsExecuted(ctx context.Context, seedName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM schema_seeds WHERE name = $1)`
	err := sr.pool.QueryRow(ctx, query, seedName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check seed execution: %w", err)
	}

	return exists, nil
}

func (sr *SeedRegistry) MarkExecuted(ctx context.Context, seedName string) error {
	query := `
		INSERT INTO schema_seeds (name, executed_at)
		VALUES ($1, NOW())
		ON CONFLICT (name) DO NOTHING
	`
	_, err := sr.pool.Exec(ctx, query, seedName)
	return err
}

func (sr *SeedRegistry) MarkUnexecuted(ctx context.Context, seedName string) error {
	query := `DELETE FROM schema_seeds WHERE name = $1`
	_, err := sr.pool.Exec(ctx, query, seedName)
	return err
}

func (sr *SeedRegistry) GetExecutedSeeds(ctx context.Context) ([]string, error) {
	query := `SELECT name FROM schema_seeds ORDER BY executed_at`
	rows, err := sr.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query executed seeds: %w", err)
	}
	defer rows.Close()

	var seeds []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan seed name: %w", err)
		}
		seeds = append(seeds, name)
	}

	return seeds, rows.Err()
}

func (sr *SeedRegistry) RunSeed(ctx context.Context, seed Seed, cfg *config.Config) error {
	executed, err := sr.IsExecuted(ctx, seed.Name)
	if err != nil {
		return fmt.Errorf("failed to check seed status: %w", err)
	}

	if executed {
		return nil
	}

	if err := seed.Up(ctx, sr.pool, cfg); err != nil {
		return fmt.Errorf("seed %s failed: %w", seed.Name, err)
	}

	if err := sr.MarkExecuted(ctx, seed.Name); err != nil {
		return fmt.Errorf("failed to mark seed as executed: %w", err)
	}

	return nil
}

func (sr *SeedRegistry) RollbackSeed(ctx context.Context, seed Seed, cfg *config.Config) error {
	executed, err := sr.IsExecuted(ctx, seed.Name)
	if err != nil {
		return fmt.Errorf("failed to check seed status: %w", err)
	}

	if !executed {
		return fmt.Errorf("seed %s not executed", seed.Name)
	}

	if seed.Down != nil {
		if err := seed.Down(ctx, sr.pool, cfg); err != nil {
			return fmt.Errorf("seed %s rollback failed: %w", seed.Name, err)
		}
	}

	if err := sr.MarkUnexecuted(ctx, seed.Name); err != nil {
		return fmt.Errorf("failed to mark seed as unexecuted: %w", err)
	}

	return nil
}

func (sr *SeedRegistry) RunSeeds(ctx context.Context, seeds []Seed, cfg *config.Config) error {
	for i, seed := range seeds {
		executed, err := sr.IsExecuted(ctx, seed.Name)
		if err != nil {
			return fmt.Errorf("failed to check seed status for %s: %w", seed.Name, err)
		}

		if executed {
			continue
		}

		if err := sr.RunSeed(ctx, seed, cfg); err != nil {
			return fmt.Errorf("seed %d/%d (%s) failed: %w", i+1, len(seeds), seed.Name, err)
		}
	}
	return nil
}

func (sr *SeedRegistry) RollbackSeeds(ctx context.Context, seeds []Seed, all bool, cfg *config.Config) error {
	executedSeeds, err := sr.GetExecutedSeeds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get executed seeds: %w", err)
	}

	if len(executedSeeds) == 0 {
		return fmt.Errorf("no executed seeds to rollback")
	}

	seedMap := make(map[string]Seed)
	for _, seed := range seeds {
		seedMap[seed.Name] = seed
	}

	var seedsToRollback []Seed
	for _, executedName := range executedSeeds {
		if seed, exists := seedMap[executedName]; exists {
			seedsToRollback = append(seedsToRollback, seed)
		}
	}

	if len(seedsToRollback) == 0 {
		return fmt.Errorf("no matching seeds found to rollback")
	}

	sort.Slice(seedsToRollback, func(i, j int) bool {
		return seedsToRollback[i].Timestamp.Before(seedsToRollback[j].Timestamp)
	})

	if all {
		for i := len(seedsToRollback) - 1; i >= 0; i-- {
			seed := seedsToRollback[i]
			if err := sr.RollbackSeed(ctx, seed, cfg); err != nil {
				return fmt.Errorf("rollback seed %d/%d (%s) failed: %w", len(seedsToRollback)-i, len(seedsToRollback), seed.Name, err)
			}
		}
	} else {
		seed := seedsToRollback[len(seedsToRollback)-1]
		if err := sr.RollbackSeed(ctx, seed, cfg); err != nil {
			return fmt.Errorf("rollback seed (%s) failed: %w", seed.Name, err)
		}
	}

	return nil
}
