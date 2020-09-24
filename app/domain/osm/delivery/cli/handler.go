package cli

import "github.com/GarinAG/gofias/domain/osm/service"

type Handler struct {
	osmService *service.OsmService
}

func NewHandler(osmService *service.OsmService) *Handler {
	return &Handler{
		osmService: osmService,
	}
}

type Node struct {
	Tags []interface{}
}

func (h *Handler) Update() {
	h.osmService.Update()
}
