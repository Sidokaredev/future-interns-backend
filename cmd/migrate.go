package cmd

import (
	initializer "future-interns-backend/init"

	"github.com/spf13/cobra"
)

var startMigration = &cobra.Command{
	Use:   "migrate",
	Short: "migrate GORM struct model into tables",
	Long: `
  there are 3 main migrations
  --target=basic      : only migrate for basic models that have'nt foreign key
  --target=candidates : perform migration for all required candidate models
  --target=employers  : perform migration for all required employee models
  `,
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		initializer.Migrate(target)
	},
}

func init() {
	rootCmd.AddCommand(startMigration)
	startMigration.Flags().StringP("target", "t", "basic", "target options to perform migration models")
}
