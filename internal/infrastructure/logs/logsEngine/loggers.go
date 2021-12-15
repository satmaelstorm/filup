package logsEngine

import (
	"bytes"
	"log"
	"os"
)

const NameForStdLogger = "STD"

const (
	LogTrace    = "trace"
	LogInfo     = "info"
	LogDebug    = "debug"
	LogError    = "error"
	LogCritical = "critical"
	LogFatal    = "fatal"
	LogProfile  = "profile"
)

type LogConfig struct {
	Prefix      string      `json:"prefix"`
	MetricsOpts CounterOpts `json:"metricsOpts"`
}

type LogConfigs map[string]LogConfig

type Loggers map[string]*log.Logger

func (l Loggers) G(name string) *log.Logger {
	logger, ok := l[name]
	if !ok {
		panic("No logger " + name)
	}
	return logger
}

func (l Loggers) Trace() *log.Logger {
	return l.G(LogTrace)
}

func (l Loggers) Debug() *log.Logger {
	return l.G(LogDebug)
}

func (l Loggers) Error() *log.Logger {
	return l.G(LogError)
}

func (l Loggers) Critical() *log.Logger {
	return l.G(LogCritical)
}

func InitLogger(project string, configs LogConfigs) (Loggers, error) {
	loggers := make(map[string]*log.Logger)
	for name, c := range configs {
		if name == NameForStdLogger {
			continue
		}

		l, err := initLogger(project, c)

		if err != nil {
			return nil, err
		}

		loggers[name] = l
	}

	c, ok := configs[NameForStdLogger]

	if ok {
		err := initStdLogger(project, c)
		if err != nil {
			return nil, err
		}
	}

	return loggers, nil
}

func initLogger(p string, c LogConfig) (*log.Logger, error) {
	return log.New(NewWriterWithCounter(os.Stderr, c.MetricsOpts), prefix(p, c.Prefix), flags()), nil
}

func initStdLogger(p string, c LogConfig) error {
	log.SetPrefix(prefix(p, c.Prefix))
	log.SetFlags(flags())

	return nil
}

func prefix(p string, pr string) string {
	buffer := new(bytes.Buffer)
	if p != "" {
		buffer.WriteString(p)
		buffer.WriteString(" ")
	}
	if pr != "" {
		buffer.WriteString("[")
		buffer.WriteString(pr)
		buffer.WriteString("] ")
	}

	return buffer.String()
}

func flags() int {
	return log.Ldate | log.Ltime | log.LUTC | log.Lmicroseconds | log.Lshortfile
}

func InitLoggersByConfig(cfg LogConfigs, projectName string, fatalIfFail bool) Loggers {
	hn, err := os.Hostname()
	if err != nil {
		log.Println("Can't get hostname")
	}

	if nil == cfg {
		cfg = make(LogConfigs)
	}

	for name, logger := range defaultLoggers {
		if _, ok := cfg[name]; !ok {
			cfg[name] = logger
		}
	}

	ls, err := InitLogger(projectName+" (host: "+hn+")", cfg)
	if fatalIfFail && err != nil {
		log.Fatalf("Error logger initialization: %s", err)
	}
	return ls
}
