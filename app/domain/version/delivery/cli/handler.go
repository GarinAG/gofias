package cli

import (
	"github.com/GarinAG/gofias/domain/version/entity"
	versionService "github.com/GarinAG/gofias/domain/version/service"
)

// Обработчик версий БД
type Handler struct {
	versionService versionService.VersionService // Сервис управления версиями
}

// Инициализация обработчика
func NewHandler(v versionService.VersionService) *Handler {
	return &Handler{
		versionService: v,
	}
}

// Получить информацию о текущей версии
func (h *Handler) GetVersionInfo() *entity.Version {
	v := h.versionService.GetLastVersionInfo()
	return v
}
