package health

import (
	"github.com/prometheus/client_golang/prometheus"
)

type HealthCollector struct {
	Health *Health

	CpuCores      *prometheus.GaugeVec
	CpuThreads    prometheus.Gauge
	CpuMaxProc    prometheus.Gauge
	CpuNumCgo     prometheus.Gauge
	CpuGoroutines prometheus.Gauge

	MemSystem        prometheus.Gauge
	MemHeap          *prometheus.GaugeVec
	MemStack         *prometheus.GaugeVec
	MemMspan         *prometheus.GaugeVec
	MemMcache        *prometheus.GaugeVec
	MemBuckcache     prometheus.Gauge
	MemOther         prometheus.Gauge
	MemGCNextTarget  prometheus.Gauge
	MemGCLast        prometheus.Gauge
	MemGCPause       prometheus.Gauge
	MemGCCount       prometheus.Gauge
	MemGCForcedCount prometheus.Gauge
}

func NewHealthCollector(namespace string, health *Health) *HealthCollector {
	res := &HealthCollector{
		Health: health,
		CpuCores: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "cpu",
			Name:      "cores",
		}, []string{"type"}),
		CpuThreads: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "cpu",
			Name:      "threads",
		}),
		CpuMaxProc: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "cpu",
			Name:      "max_proc",
		}),
		CpuNumCgo: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "cpu",
			Name:      "cgo_calls_count",
		}),
		CpuGoroutines: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "cpu",
			Name:      "goroutines_count",
		}),

		MemSystem: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "system",
		}),
		MemHeap: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "heap",
		}, []string{"type"}),
		MemStack: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "stack",
		}, []string{"type"}),
		MemMspan: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "mspan",
		}, []string{"type"}),
		MemMcache: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "mcache",
		}, []string{"type"}),
		MemBuckcache: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "buckcache",
		}),
		MemOther: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "other",
		}),
		MemGCNextTarget: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "gc_next_target",
		}),
		MemGCLast: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "gc_last_time",
		}),
		MemGCPause: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "gc_pause_sum",
		}),
		MemGCCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "gc_count",
		}),
		MemGCForcedCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "mem",
			Name:      "gc_forced_count",
		}),
	}

	prometheus.MustRegister(res)

	return res
}

func (collector *HealthCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(collector, ch)
}

func (collector *HealthCollector) Collect(ch chan<- prometheus.Metric) {
	collector.Health.UpdateHealth()

	collector.CpuCores.Collect(ch)
	collector.CpuThreads.Collect(ch)
	collector.CpuMaxProc.Collect(ch)
	collector.CpuNumCgo.Collect(ch)
	collector.CpuGoroutines.Collect(ch)

	collector.MemSystem.Collect(ch)
	collector.MemHeap.Collect(ch)
	collector.MemStack.Collect(ch)
	collector.MemMspan.Collect(ch)
	collector.MemMcache.Collect(ch)
	collector.MemBuckcache.Collect(ch)
	collector.MemOther.Collect(ch)
	collector.MemGCNextTarget.Collect(ch)
	collector.MemGCLast.Collect(ch)
	collector.MemGCPause.Collect(ch)
	collector.MemGCCount.Collect(ch)
	collector.MemGCForcedCount.Collect(ch)
}
