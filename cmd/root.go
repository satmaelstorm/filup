package cmd

import (
	"github.com/pkg/errors"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strconv"
)

var ( //set by go build -ldflags '-s -w -X "cmd.BuildTime=${VERSION}" -X "cmd.GitBranch=${BRANCH}" -X "cmd.GitCommit=${COMMIT}"
	BuildTime string
	GitBranch string
	GitCommit string
)

var (
	cfgName string
	rootCmd = &cobra.Command{
		Use:   config.ProjectName,
		Short: config.ProjectName + " service",
		Long:  "File Upload service - upload files directly to S3-compatibility storage. Supports multipart/form-data and websockets. Cloud ready.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return loadConfig(cfgName)
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgName, "config", "", "config file")
	rootCmd.AddCommand(serverCmd)
}

func loadConfig(configName string) error {
	log.Println("pid: " + strconv.Itoa(os.Getpid()))
	log.Println("build: " + BuildTime)
	cfg, err := config.LoadConfigByViper(configName)
	if err != nil {
		return errors.Wrap(err, "error while load configuration")
	}
	logs.ProvideLoggers(cfg)
	return err
}
