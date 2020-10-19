package cli

import (
	service2 "github.com/GarinAG/gofias/domain/address/service"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	"github.com/GarinAG/gofias/interfaces"
)

// Обработчик основного импорта
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

// Проверка обновлений
func (h *Handler) CheckUpdates(fiasApi *fiasApiService.FiasApiService, versionService *versionService.VersionService) {
	// Получает последнюю загруженную версию
	v := versionService.GetLastVersionInfo()
	h.logger.WithFields(interfaces.LoggerFields{
		"version": v,
	}).Debug("Last version info")

	if v != nil {
		// Загрузка дельт
		h.importService.StartDeltaImport(fiasApi, versionService, v)
	} else {
		// Загрузка полного импорта
		h.importService.StartFullImport(fiasApi, versionService)
	}
	// Обновление индексов
	h.importService.Index()
}
