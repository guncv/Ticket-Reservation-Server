package seed

import (
	_ "github.com/guncv/ticket-reservation-server/internal/db/seeders"
	"github.com/spf13/cobra"
)

var RollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback executed seeders",
	Long: `Rollback executed seeders in reverse chronological order.

By default, rolls back only the most recent seeder. Use --all flag to rollback
all executed seeders.

Rollbacks execute in reverse order (newest first, oldest last) to ensure
dependencies are handled correctly. Each rollback removes the seed from the
schema_seeds tracking table.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RollbackCmd.Flags().Bool("all", false, "revert all executed seeders")
}

// func runRollback(cmd *cobra.Command, args []string) error {
// 	all, _ := cmd.Flags().GetBool("all")
// 	action := "rollback"
// 	if all {
// 		action = "rollback all"
// 	}
// 	fmt.Println("seed " + action + " started")

// 	seeds, err := db.LoadSeeds()
// 	if err != nil {
// 		return fmt.Errorf("failed to load seeds: %w", err)
// 	}

// 	if len(seeds) == 0 {
// 		fmt.Println("no seeds found")
// 		return nil
// 	}

// 	dbPool, err := db.NewPgPool(cfg)
// 	if err != nil {
// 		return fmt.Errorf("failed to create database connection: %w", err)
// 	}

// 	seedRegistry := db.NewSeedRegistry(dbPool.Pool)
// 	ctx := context.Background()

// 	if err := seedRegistry.RollbackSeeds(ctx, seeds, all); err != nil {
// 		return fmt.Errorf("seed rollback failed: %w", err)
// 	}

// 	fmt.Println("seed " + action + " completed successfully")
// 	return nil
// }
