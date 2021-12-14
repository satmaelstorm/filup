package config

import (
	"github.com/satmaelstorm/envviper"
	"github.com/satmaelstorm/filup/internal/domain/port"
)

var gConfig Configuration

func ProvideConfig() Configuration {
	return gConfig
}

func ProvideUploaderConfig() port.UploaderConfig {
	c := gConfig.Uploader
	return c
}

func LoadConfigByViper(name string) (Configuration, error) {
	viper := envviper.NewEnvViper()

	if name != "" {
		viper.AddConfigPath("conf/")
		viper.AddConfigPath("/etc/bear/")
		viper.AddConfigPath(".")
		viper.SetConfigName(name)
		viper.SetConfigType("yaml")
		err := viper.ReadInConfig()
		if err != nil {
			return Configuration{}, err
		}
	} else {
		err := defaultConfig(viper)
		if err != nil {
			return Configuration{}, err
		}
	}
	var cfg Configuration
	viper.SetEnvParamsSimple(ProjectName)
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return Configuration{}, err
	}

	cfg.AfterLoad()
	gConfig = cfg

	return cfg, nil
}
