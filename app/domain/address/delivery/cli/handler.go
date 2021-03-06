package cli

import (
	service2 "github.com/GarinAG/gofias/domain/address/service"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	osmService "github.com/GarinAG/gofias/domain/osm/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	"github.com/GarinAG/gofias/interfaces"
)

// Обработчик основного импорта
type Handler struct {
	importService *service2.ImportService    // Сервис импорта
	osmService    *osmService.OsmService     // Сервис OSM
	logger        interfaces.LoggerInterface // Логгер
}

// Инициализация обработчика
func NewHandler(s *service2.ImportService, osm *osmService.OsmService, logger interfaces.LoggerInterface) *Handler {
	return &Handler{
		importService: s,
		osmService:    osm,
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
	// Обновление гео-данных
	if !h.importService.SkipOsm {
		h.osmService.Update()
	}
}
