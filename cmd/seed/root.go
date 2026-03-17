package seed

import (
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	cfgErr error
)

var RootCmd = &cobra.Command{
	Use:   "seeder",
	Short: "Database seeder CLI",
	Long: `Database seeder CLI for managing seed data.

Seeders work like migrations - they run in chronological order based on their
filename timestamps. Each seeder can have both an "up" function (to seed data)
and a "down" function (to rollback the seed).

Seed files are automatically discovered from the db/seeders directory and must
follow the naming convention: YYYYMMDDHHMMSS_seed_name.go`,
	PersistentPreRunE: func(c *cobra.Command, args []string) (err error) {
		cfg, cfgErr = config.LoadConfig(c.Root().PersistentFlags())
		if cfgErr != nil {
			return cfgErr
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(newSeedCmd)
	RootCmd.AddCommand(seedCmd)
	// RootCmd.AddCommand(RollbackCmd)
}
