package handler

import (
	"context"
	"github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/v1/health"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"math"
	"runtime"
	"time"
)

const MB float64 = 1.0 * 1024 * 1024

type HealthHandler struct {
	Server *grpc.Server
	start  time.Time
}

func NewHealthHandler() *HealthHandler {
	handler := &HealthHandler{
		start: time.Now(),
	}

	return handler
}

func (h *HealthHandler) CheckHealth(ctx context.Context, empty *empty.Empty) (*health.Health, error) {
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)

	return &health.Health{
		Uptime:               h.GetUptime(),
		AllocatedMemory:      h.toMegaBytes(mem.Alloc),
		TotalAllocatedMemory: h.toMegaBytes(mem.TotalAlloc),
		Goroutines:           int32(runtime.NumGoroutine()),
		NumberOfCPUs:         int32(runtime.NumCPU()),
		GCCycles:             mem.NumGC,
		HeapSys:              h.toMegaBytes(mem.HeapSys),
		HeapAllocated:        h.toMegaBytes(mem.HeapAlloc),
		ObjectsInUse:         mem.Mallocs - mem.Frees,
		OSMemoryObtained:     h.toMegaBytes(mem.Sys),
	}, nil
}

func (h *HealthHandler) GetUptime() int64 {
	return time.Now().Unix() - h.start.Unix()
}

func (h *HealthHandler) toMegaBytes(bytes uint64) float32 {
	return h.toFixed(float64(bytes)/MB, 2)
}

func (h *HealthHandler) round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func (h *HealthHandler) toFixed(num float64, precision int) float32 {
	output := math.Pow(10, float64(precision))
	return float32(h.round(num*output)) / float32(output)
}
