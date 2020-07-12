package cli

import (
	service2 "github.com/GarinAG/gofias/domain/address/service"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	"github.com/GarinAG/gofias/interfaces"
)

type Handler struct {
	importService *service2.ImportService
	logger        interfaces.LoggerInterface
}

func NewHandler(s *service2.ImportService, logger interfaces.LoggerInterface) *Handler {
	return &Handler{
		importService: s,
		logger:        logger,
	}
}

func (h *Handler) CheckUpdates(fiasApi *fiasApiService.FiasApiService, versionService *versionService.VersionService) {
	v := versionService.GetLastVersionInfo()

	if v != nil {
		h.importService.CheckUpdates(fiasApi, versionService, v)
	} else {
		h.importService.StartFullImport(fiasApi, versionService)
	}
	h.importService.Index()
}
