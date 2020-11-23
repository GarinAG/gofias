package cli

import (
	grpc2 "github.com/GarinAG/gofias/infrastructure/persistence/grpc"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
)

// Обработчик GRPC
type Handler struct {
	ctn *registry.Container // Контейнер
}

// Инициализация обработчика
func NewHandler(ctn *registry.Container) *Handler {
	return &Handler{
		ctn: ctn,
	}
}

// Запускает grpc-сервер
func (h *Handler) Run() {
	app := grpc2.NewGrpcServer(h.ctn)
	if err := app.Run(); err != nil {
		app.Logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Program fatal error")
	}
}
