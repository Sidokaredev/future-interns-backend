package cmd

import (
	"database-migration-cli/internal/migrations"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Performs a migration for database tables",
	Long: `
    Migrate the entire available model into database tables.
    NOTE: This command will drop the table if exist before create the new table.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		migrations.FreshMigration()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
