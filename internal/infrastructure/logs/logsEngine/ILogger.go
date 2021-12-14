package logsEngine

import "log"

type ILogger interface {
	Trace() *log.Logger
	Debug() *log.Logger
	Error() *log.Logger
	Critical() *log.Logger
}
