package config

import (
	"bytes"
	_ "embed"
	"github.com/satmaelstorm/envviper"
)

//go:embed default.yaml
var defConfig []byte

func defaultConfig(vp *envviper.EnvViper) error {
	vp.SetConfigType("yaml")
	return vp.ReadConfig(bytes.NewBuffer(defConfig))
}
