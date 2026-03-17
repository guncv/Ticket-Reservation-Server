package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
)

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "run migrations up",
	Long:  `All pending migrations will be applied.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := mg.MigrateUp()
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		return nil
	},
}
