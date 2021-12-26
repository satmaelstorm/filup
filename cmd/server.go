package cmd

import (
	"github.com/satmaelstorm/filup/internal/infrastructure/di"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start Filup-server",
	Long:  "Start Filup-server",
	RunE: func(cmd *cobra.Command, args []string) error {
		server, err := di.InitWebServer()
		if err != nil {
			return err
		}
		server.Run()
		return nil
	},
}
