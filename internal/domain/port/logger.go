package port

import "log"

type Logger interface {
	Critical() *log.Logger
	Error() *log.Logger
	Trace() *log.Logger
	Debug() *log.Logger
}
