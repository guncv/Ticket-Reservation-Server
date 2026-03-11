package seed

import (
	"fmt"

	"github.com/guncv/ticket-reservation-server/internal/db/seeders"
	"github.com/spf13/cobra"
)

var newSeedCmd = &cobra.Command{
	Use:   "new <seeder-name>",
	Short: "Create a new seeder file",
	Long: `Create a new seeder file with the specified name.

The seeder name should follow these conventions:
- Use snake_case format (lowercase with underscores)
- Be descriptive of what data is being seeded
- Examples: admin_account, customer_data, product_catalog

The command will automatically:
- Prefix the file with a timestamp (YYYYMMDDHHMMSS)
- Create a seeder function template with both Up and Down functions
- Place files in the db/seeders directory
- Register the seed automatically via init() function

The generated file will follow the naming convention:
YYYYMMDDHHMMSS_seed_<seeder-name>.go`,
	Example: `
# Create a new seeder for admin account
seeder new admin_account

# Create a seeder for customer data
seeder new customer_data

# Create a seeder for product catalog
seeder new product_catalog`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("seeder name is required")
		}

		if err := seeders.NewSeeder(args[0], cfg); err != nil {
			return fmt.Errorf("failed to create seeder: %w", err)
		}

		return nil
	},
}
