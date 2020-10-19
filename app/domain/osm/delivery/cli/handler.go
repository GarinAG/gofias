package cli

import "github.com/GarinAG/gofias/domain/osm/service"

// Обработчик разбора OpenStreetMap
type Handler struct {
	osmService *service.OsmService // Сервис работы с OSM
}

// Инициализация обработчика
func NewHandler(osmService *service.OsmService) *Handler {
	return &Handler{
		osmService: osmService,
	}
}

// Обновляет данные местоположений
func (h *Handler) Update() {
	h.osmService.Update()
}
