package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
)

var newMigrationCmd = &cobra.Command{
	Use:   "new <migration-name>",
	Short: "create a new migration",
	Long: `
Create a new migration file with the specified name.

The migration name should follow these conventions:
- Use snake_case format (lowercase with underscores)
- Start with an action verb (create, add, remove, alter, etc.)
- Be descriptive of the change
- Include the table name

The command will automatically:
- Prefix the file with a timestamp
- Create up and down migration files
- Place files in the migrations directory`,
	Example: `
# Create a new migration for users table
lms migrate new create_users_table

# Add a column to existing table
lms migrate new add_email_to_users

# Create a join table
lms migrate new create_user_roles_table

# Modify existing table
lms migrate new alter_users_add_timestamps`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Error: Migration name is required")
			fmt.Println("Usage: lms migrate <migration-name>")
			return
		}

		err := mg.NewMigrate(args[0])
		if err != nil {
			fmt.Printf("Error creating migration: %v\n", err)
			return
		}

		fmt.Printf("Successfully created migration: %s\n", args[0])
	},
}
