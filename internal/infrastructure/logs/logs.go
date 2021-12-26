package logs

import (
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
)

var ls logsEngine.Loggers

func ProvideLoggers(cfg config.Configuration) *logsEngine.Loggers {
	if nil == ls {
		ls = logsEngine.InitLoggersByConfig(cfg.Logs, config.ProjectName, true)
	}
	return &ls
}
