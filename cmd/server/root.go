package server

import (
	"fmt"
	"os"

	"github.com/guncv/ticket-reservation-server/cmd/migrate"
	"github.com/guncv/ticket-reservation-server/cmd/seed"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/containers"
	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	cfgErr error
)

var rootCmd = &cobra.Command{
	Use:   "ticket-reservation-server",
	Short: "Ticket Reservation Server is an API for managing ticket reservations",
	Long: `Ticket Reservation Server is a backend system for managing ticket reservations.

This application provides:
- User management and authentication
- Ticket reservation management
- RESTful API endpoints
- Configurable settings via environment variables, config files, or command-line flags

The system supports multiple environments (development, production) and can be configured
through YAML configuration files, environment variables, or command-line flags with the
following precedence: defaults < config file < environment variables < command-line flags.

Examples:
	# Run with default configuration
	ticket-reservation-server

	# Run on a custom port
	ticket-reservation-server --app-port 3000

	# Run database migrations
	ticket-reservation-server migrate`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, cfgErr = config.LoadConfig(cmd.Root().PersistentFlags())
		if cfgErr != nil {
			return cfgErr
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		c := containers.NewContainer(cfg)
		if err := c.Run().Error; err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(migrate.RootCmd)
	rootCmd.AddCommand(seed.RootCmd)

	rootCmd.PersistentFlags().String("app-port", "", "application port")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
