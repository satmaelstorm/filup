package health

import (
	"github.com/klauspost/cpuid"
	"os"
	"runtime"
	"sync"
	"time"
)

const DefaultMemoryUpdatePeriod = time.Duration(60) * time.Second

type Health struct {
	Pid                int           `json:"pid"`
	TimeStart          int64         `json:"started_at"`
	Host               string        `json:"host_name"`
	Memory             MemoryHealth  `json:"memory"`
	CpuHealth          CpuHealth     `json:"cpu"`
	Ready              bool          `json:"ready"`
	LastMemoryUpdate   int64         `json:"last_memory_update"`
	MemoryUpdatePeriod time.Duration `json:"memory_update_period"`

	collector *HealthCollector
	mu        sync.Mutex
}

type CpuHealth struct {
	CpuCores           int   `json:"cpu_physical_cores"`
	CpuThreadsPerCores int   `json:"cpu_threads_per_cores"`
	CpuLogicalCores    int   `json:"cpu_logical_cores"`
	MaxProc            int   `json:"go_max_proc"`
	NumCgo             int64 `json:"num_cgo_call"`
	NumGoroutine       int   `json:"num_goroutine"`
}

type MemoryHealth struct {
	SystemMemory  uint64 `json:"system_memory"`
	HeapAllocated uint64 `json:"heap_allocated"`
	StackInuse    uint64 `json:"stack_inuse"`
	MSpanInuse    uint64 `json:"mspan_inuse"`
	MCacheInuse   uint64 `json:"mcache_inuse"`
	HeapInuse     uint64 `json:"heap_inuse"`
	HeapSys       uint64 `json:"heap_sys"`
	StackSys      uint64 `json:"stack_sys"`
	MSpanSys      uint64 `json:"mspan_sys"`
	MCacheSys     uint64 `json:"mcache_sys"`
	BuckHashSys   uint64 `json:"buckcache_sys"`
	OtherSys      uint64 `json:"other_sys"`

	LastGC      uint64 `json:"last_gc"`
	NextGC      uint64 `json:"next_gc"`
	PauseGC     uint64 `json:"pause_gc"`
	NumGC       uint32 `json:"num_gc"`
	NumForcedGC uint32 `json:"num_forced_gc"`
}

var timeStart time.Time

func init() {
	timeStart = time.Now()
}

func NewHealth() *Health {
	hn, _ := os.Hostname()
	res := &Health{
		Pid:                os.Getpid(),
		Host:               hn,
		TimeStart:          timeStart.Unix(),
		MemoryUpdatePeriod: DefaultMemoryUpdatePeriod,
	}
	return res
}

func (h *Health) CurrentHealth() {
	h.UpdateHealth()
}

func (h *Health) UpdateHealth() {
	h.mu.Lock()
	defer h.mu.Unlock()

	timeDiff := time.Now().UnixNano() - h.LastMemoryUpdate
	if timeDiff < int64(h.MemoryUpdatePeriod) {
		return
	}

	h.CpuHealth.CpuCores = cpuid.CPU.PhysicalCores
	h.CpuHealth.CpuLogicalCores = cpuid.CPU.LogicalCores
	h.CpuHealth.CpuThreadsPerCores = cpuid.CPU.ThreadsPerCore
	h.CpuHealth.MaxProc = runtime.GOMAXPROCS(0)
	h.CpuHealth.NumCgo = runtime.NumCgoCall()
	h.CpuHealth.NumGoroutine = runtime.NumGoroutine()
	h.updateCPUMetrics()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	h.Memory.SystemMemory = m.Sys
	h.Memory.HeapAllocated = m.HeapAlloc
	h.Memory.StackInuse = m.StackInuse
	h.Memory.MSpanInuse = m.MSpanInuse
	h.Memory.MCacheInuse = m.MCacheInuse
	h.Memory.HeapInuse = m.HeapInuse
	h.Memory.HeapSys = m.HeapSys
	h.Memory.StackSys = m.StackSys
	h.Memory.MSpanSys = m.MSpanSys
	h.Memory.BuckHashSys = m.BuckHashSys
	h.Memory.OtherSys = m.OtherSys
	h.Memory.LastGC = m.LastGC
	h.Memory.NextGC = m.NextGC
	h.Memory.NumForcedGC = m.NumForcedGC
	h.Memory.NumGC = m.NumGC
	h.Memory.PauseGC = m.PauseTotalNs
	h.updateMemoryMetrics()

	h.LastMemoryUpdate = time.Now().UnixNano()
}

func (h *Health) SetReady(r bool) {
	h.Ready = r
}

func (h *Health) RegisterCollector(namespace string) {
	NewHealthCollector(namespace, h)

	h.CpuHealth.CpuThreadsPerCores = cpuid.CPU.ThreadsPerCore
}

func (h *Health) updateCPUMetrics() {
	if h.collector == nil {
		return
	}

	h.collector.CpuCores.WithLabelValues("logical").Set(float64(h.CpuHealth.CpuLogicalCores))
	h.collector.CpuCores.WithLabelValues("physical").Set(float64(h.CpuHealth.CpuCores))
	h.collector.CpuThreads.Set(float64(h.CpuHealth.CpuThreadsPerCores))
	h.collector.CpuMaxProc.Set(float64(h.CpuHealth.MaxProc))
	h.collector.CpuNumCgo.Set(float64(h.CpuHealth.NumCgo))
	h.collector.CpuGoroutines.Set(float64(h.CpuHealth.NumGoroutine))
}

func (h *Health) updateMemoryMetrics() {
	if h.collector == nil {
		return
	}
	mem := h.Memory

	h.collector.MemSystem.Set(float64(mem.SystemMemory))

	h.collector.MemHeap.WithLabelValues("alloc").Set(float64(mem.HeapAllocated))
	h.collector.MemHeap.WithLabelValues("inuse").Set(float64(mem.HeapInuse))
	h.collector.MemHeap.WithLabelValues("sys").Set(float64(mem.HeapSys))

	h.collector.MemStack.WithLabelValues("inuse").Set(float64(mem.StackInuse))
	h.collector.MemStack.WithLabelValues("sys").Set(float64(mem.StackSys))

	h.collector.MemMspan.WithLabelValues("inuse").Set(float64(mem.MSpanInuse))
	h.collector.MemMspan.WithLabelValues("sys").Set(float64(mem.MSpanSys))

	h.collector.MemMcache.WithLabelValues("inuse").Set(float64(mem.MCacheInuse))
	h.collector.MemMcache.WithLabelValues("sys").Set(float64(mem.MCacheSys))

	h.collector.MemBuckcache.Set(float64(mem.BuckHashSys))
	h.collector.MemOther.Set(float64(mem.OtherSys))
	h.collector.MemGCNextTarget.Set(float64(mem.NextGC))
	h.collector.MemGCLast.Set(float64(mem.LastGC))
	h.collector.MemGCPause.Set(float64(mem.PauseGC))
	h.collector.MemGCCount.Set(float64(mem.NumGC))
	h.collector.MemGCForcedCount.Set(float64(mem.NumForcedGC))
}

// inlining
func boolToFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
}
