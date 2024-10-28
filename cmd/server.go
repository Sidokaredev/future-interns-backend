package cmd

import (
	"future-interns-backend/internal"

	"github.com/spf13/cobra"
)

var startServer = &cobra.Command{
	Use:   "serve",
	Short: "run gin.Engine instance",
	Run: func(cmd *cobra.Command, args []string) {
		address, _ := cmd.Flags().GetString("address")
		internal.CreateServer(address) // start Gin htpp server
	},
}

func init() {
	rootCmd.AddCommand(startServer)
	startServer.Flags().StringP("address", "a", ":3000", "specify the address to gin.Engine when it begin to run")
}
