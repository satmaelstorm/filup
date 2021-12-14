package logsEngine

import (
	"github.com/prometheus/client_golang/prometheus"
	"io"
)

type WriterWithCounter struct {
	writer  io.Writer
	counter prometheus.Counter
}

func (w *WriterWithCounter) Write(p []byte) (n int, err error) {
	if w.counter != nil {
		w.counter.Inc()
	}
	return w.writer.Write(p)
}

func NewWriterWithCounter(writer io.Writer, opts CounterOpts) *WriterWithCounter {
	w := new(WriterWithCounter)
	w.writer = writer
	w.counter = nil
	if "" == opts.Namespace || "" == opts.Name {
		return w
	}
	subsystem := opts.Subsystem
	if "" == subsystem {
		subsystem = "logs"
	}
	w.counter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: opts.Namespace,
		Subsystem: subsystem,
		Name:      opts.Name,
	})
	prometheus.MustRegister(w.counter)
	return w
}
