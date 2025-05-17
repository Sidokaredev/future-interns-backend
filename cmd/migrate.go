package cmd

import (
	initializer "future-interns-backend/init"
	"log"

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

var seedUser = &cobra.Command{
	Use:   "seed",
	Short: "seed some user for testing purposes",
	Long: `
		there are 4 main user seed
		--for=candidate
		--for=employer
		--for=university
		--for=administrator
		you have to provide 2 Flags, --email and --password. that are completely required
	`,
	Run: func(cmd *cobra.Command, args []string) {
		seedFor, errSeedFor := cmd.Flags().GetString("target")
		email, errEmail := cmd.Flags().GetString("email")
		password, errPassword := cmd.Flags().GetString("password")
		if errSeedFor != nil || errEmail != nil || errPassword != nil {
			log.Fatalf("invalid command flags. for: %v, email: %v, password: %v", errSeedFor.Error(), errEmail.Error(), errPassword.Error())
			return
		}

		initializer.SeedUser(seedFor, email, password)
	},
}

/* Final Code */
// var tableMigration = &cobra.Command{
// 	Use:   "dbmigrate",
// 	Short: "Melakukan migrasi database yang telah didefinisikan melalui model pada direktori ./internal/models/*",
// 	Long: `

// 	`,
// }

func init() {
	rootCmd.AddCommand(startMigration, seedUser)
	startMigration.Flags().StringP("target", "t", "basic", "target options to perform migration models")
	seedUser.Flags().StringP("target", "t", "", "user seed target candidate, employer or administrator")
	seedUser.Flags().StringP("email", "m", "", "email for user")
	seedUser.Flags().StringP("password", "p", "", "password for user")
}
