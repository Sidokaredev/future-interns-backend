package cmd

import (
	initializer "database-migration-cli/init"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "sdkrdev",
	Short: "A CLI to performs all migrations in Future Interns Web Applications",
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("Everything is ready to migrate")
	// },
}

func init() {
	cobra.OnInitialize(InitConfig, initializer.GormSQLServer)
}

func Execute() {
	if errExecuteCobra := rootCmd.Execute(); errExecuteCobra != nil {
		panic(errExecuteCobra)
	}
}

func InitConfig() {
	viper.SetConfigName("sql-server-db")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	if errReadConfig := viper.ReadInConfig(); errReadConfig != nil {
		panic(errReadConfig)
	}

	viper.AutomaticEnv()
	fmt.Println("viper status \t:", viper.GetString("status"))
}
