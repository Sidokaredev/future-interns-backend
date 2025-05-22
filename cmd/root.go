package cmd

import (
	initializer "future-interns-backend/init"
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sdkrdev",
	Short: "A CLI command to initiate Future Interns project",
}

func init() {
	initializer.LoadAppConfig()            // load ./configs/config.yaml
	err := initializer.GormSQLServerInit() // open connection to sql server with GORM
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	// initializer.DockerClientInit() // control docker over socker --mount /var/run/docker.sock
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
