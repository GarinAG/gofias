package cli

import (
	service2 "github.com/GarinAG/gofias/domain/address/service"
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

func (h *Handler) Index() {
	h.importService.Index()
}
