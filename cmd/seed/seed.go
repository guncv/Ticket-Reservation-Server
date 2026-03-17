package seed

import (
	"context"
	"fmt"

	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/guncv/ticket-reservation-server/internal/db/seeders"
	_ "github.com/guncv/ticket-reservation-server/internal/db/seeders"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Run all pending seeders",
	Long: `Run all pending seeders in chronological order within a single transaction.

Seeds are automatically discovered from the db/seeders directory and sorted
by their timestamp. Only seeds that haven't been executed yet will be run.

All seeds run atomically - if any seed fails, the entire operation is rolled back.
This ensures no partial state is left in the database.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return runSeed()
	},
}

func runSeed() error {
	fmt.Println("seed started")

	seeds, err := seeders.LoadSeeds()
	if err != nil {
		return fmt.Errorf("failed to load seeds: %w", err)
	}

	if len(seeds) == 0 {
		fmt.Println("no seeds to run")
		return nil
	}

	fmt.Println("loaded seeds:", len(seeds))
	for i, seed := range seeds {
		fmt.Println("  ", i+1, seed.Name, seed.Timestamp.Format("2006-01-02 15:04:05"))
	}

	dbPool, err := db.NewPgPool(cfg)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}
	defer dbPool.Pool.Close()

	seedRegistry := db.NewSeedRegistry(dbPool.Pool)
	ctx := context.Background()

	executedSeeds, err := seedRegistry.GetExecutedSeeds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get executed seeds: %w", err)
	}

	if len(executedSeeds) > 0 {
		fmt.Println("already executed:", len(executedSeeds))
		for _, name := range executedSeeds {
			fmt.Println("  -", name)
		}
	}

	fmt.Println("running seeds in transaction...")
	if err := seedRegistry.RunSeeds(ctx, seeds, cfg); err != nil {
		return fmt.Errorf("seed failed (rolled back): %w", err)
	}

	fmt.Println("seed completed successfully")
	return nil
}
