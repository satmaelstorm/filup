package logsEngine

var defaultLoggers = LogConfigs{
	LogTrace:    LogConfig{Prefix: "TRACE"},
	LogDebug:    LogConfig{Prefix: "DEBUG"},
	LogError:    LogConfig{Prefix: "ERROR"},
	LogCritical: LogConfig{Prefix: "CRITICAL"},
}

func InitLoggersEmpty(projectName string) Loggers {
	return InitLoggersByConfig(defaultLoggers, projectName, true)
}
