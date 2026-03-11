package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
)

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "run migrations down",
	Long:  `Revert the most recent migration. Use --all flag to revert all migrations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if all {
			err := mg.MigrateDownAll()
			if err != nil {
				return fmt.Errorf("migration down all failed: %w", err)
			}
		} else {
			err := mg.MigrateDown()
			if err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}
		}

		return nil
	},
}

func init() {
	migrateDownCmd.Flags().Bool("all", false, "revert all migrations")
}
