package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/health"
)

var (
	StdBuckets     = []float64{.01, .05, .1, .2, .3, .4, .5, .6, .7, .8, .9, 1, 2, 3}
	StdHttpBuckets = []float64{.01, .1, 0.25, .5, 0.75, 1, 3, 5, 7, 10, 15, 20, 25, 30}
)

var gHealth *health.Health

func init() {
	gHealth = health.NewHealth()
	gHealth.RegisterCollector(config.ProjectMetricsNamespace)
	prometheus.Unregister(collectors.NewGoCollector())
}
