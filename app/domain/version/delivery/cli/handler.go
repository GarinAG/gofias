package cli

import (
	versionService "github.com/GarinAG/gofias/domain/version/service"
	"log"
)

type Handler struct {
	versionService versionService.VersionService
}

func NewHandler(v versionService.VersionService) *Handler {
	return &Handler{
		versionService: v,
	}
}

func (h *Handler) GetVersionInfo() {
	v := h.versionService.GetLastVersionInfo()
	log.Println(v)
}
