package cmd

import (
	"github.com/satmaelstorm/filup/internal/infrastructure/di"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start Filup-server",
	Long:  "Start Filup-server",
	Run: func(cmd *cobra.Command, args []string) {
		server := di.InitWebServer()
		server.Run()
	},
}
