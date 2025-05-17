package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var dbmigrate = &cobra.Command{
	Use:   "dbmigrate",
	Short: "Melakukan migrasi database yang telah didefinisikan melalui model pada direktori ./internal/models/*",
	Long: `
    ARGS
    - init : Generate tabel database berdasarkan model pada direktori /internal/models/*. NOTE: lakukan secara hati-hati karena ini akan menghapus/reset seluruh rows data.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("these are arguments \t: ", args)
	},
}

func init() {
	rootCmd.AddCommand(dbmigrate)
}
