package cli

import (
	service2 "github.com/GarinAG/gofias/domain/address/service"
	"github.com/GarinAG/gofias/interfaces"
)

// Обработчик индексации БД
type Handler struct {
	importService *service2.ImportService    // Сервис импорта
	logger        interfaces.LoggerInterface // Логгер
}

// Инициализация обработчика
func NewHandler(s *service2.ImportService, logger interfaces.LoggerInterface) *Handler {
	return &Handler{
		importService: s,
		logger:        logger,
	}
}

// Индексация БД
func (h *Handler) Index() {
	h.importService.Index()
}
