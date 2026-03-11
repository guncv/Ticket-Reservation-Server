package migrate

import (
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/spf13/cobra"
)

var (
	cfg        *config.Config
	cfgErr     error
	mg         *db.Migrate
	migrateErr error
)

var RootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database schema migrations",
	Long: `The migrate command provides tools for managing database schema migrations.

	Database migrations allow you to version control your database schema and apply
	incremental changes to your database structure over time. Migrations are stored
	as SQL files in the internal/db/migrations/ directory and are automatically
	detected and executed in chronological order based on their timestamp prefix.

	This command supports the following operations:
	- Apply pending migrations to update your database schema
	- Rollback the last applied migration
	- Rollback all migrations to reset the database

	Migrations are tracked in the database, ensuring that each migration is only
	applied once and can be safely rolled back when needed.

	Examples:
		# View available migration subcommands
		ticket-reservation-server migrate

		# Apply all pending migrations
		ticket-reservation-server migrate up

		# Rollback the last migration
		ticket-reservation-server migrate down`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, cfgErr = config.LoadConfig(cmd.Root().PersistentFlags())
		if cfgErr != nil {
			return cfgErr
		}

		mg, migrateErr = db.LoadMigrate(cfg)
		if migrateErr != nil {
			return migrateErr
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(newMigrationCmd)
	RootCmd.AddCommand(migrateUpCmd)
	RootCmd.AddCommand(migrateDownCmd)
}
