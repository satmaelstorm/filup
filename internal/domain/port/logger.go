package port

import "log"

type CriticalLogger interface {
	Critical() *log.Logger
}
