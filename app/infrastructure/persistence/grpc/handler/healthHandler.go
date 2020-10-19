package handler

import (
	"context"
	"github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/health"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"math"
	"runtime"
	"time"
)

const MB float64 = 1.0 * 1024 * 1024

// GRPC-обработчик проверки состояния сервиса
type HealthHandler struct {
	Server *grpc.Server // GRPC-сервер
	start  time.Time    // Время старта сервера
}

// Инициализация обработчика
func NewHealthHandler() *HealthHandler {
	handler := &HealthHandler{
		start: time.Now(),
	}

	return handler
}

// Проверка состояния сервера
func (h *HealthHandler) CheckHealth(ctx context.Context, empty *empty.Empty) (*health.Health, error) {
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)

	return &health.Health{
		Uptime:               h.GetUptime(),                 // Время работы сервера
		AllocatedMemory:      h.toMegaBytes(mem.Alloc),      // Текущее количество потребляемой памяти
		TotalAllocatedMemory: h.toMegaBytes(mem.TotalAlloc), // Общее количество потребляемой памяти
		Goroutines:           int32(runtime.NumGoroutine()), // Общее количество процессов
		NumberOfCPUs:         int32(runtime.NumCPU()),       // Общее количество процессоров
		GCCycles:             mem.NumGC,                     // Количество завершенных циклов запуска сборщика мусора
		HeapSys:              h.toMegaBytes(mem.HeapSys),    // Количество байт памяти, переданных системой
		HeapAllocated:        h.toMegaBytes(mem.HeapAlloc),  // Количество используемой памяти
		ObjectsInUse:         mem.Mallocs - mem.Frees,
		OSMemoryObtained:     h.toMegaBytes(mem.Sys),
	}, nil
}

// Получить время работы сервера
func (h *HealthHandler) GetUptime() int64 {
	return time.Now().Unix() - h.start.Unix()
}

// Конверитровать байты в мегабайты
func (h *HealthHandler) toMegaBytes(bytes uint64) float32 {
	return h.toFixed(float64(bytes)/MB, 2)
}

// Округлить число
func (h *HealthHandler) round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// Конверитровать число в float32
func (h *HealthHandler) toFixed(num float64, precision int) float32 {
	output := math.Pow(10, float64(precision))
	return float32(h.round(num*output)) / float32(output)
}
