package handler

import (
	"context"
	"github.com/GarinAG/gofias/domain/version/service"
	"github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/version"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"strconv"
)

type VersionHandler struct {
	Version        string
	Server         *grpc.Server
	versionService *service.VersionService
}

func NewVersionHandler(a *service.VersionService, v string) *VersionHandler {
	handler := &VersionHandler{
		Version:        v,
		versionService: a,
	}

	return handler
}

func (h *VersionHandler) GetVersion(ctx context.Context, empty *empty.Empty) (*version.Version, error) {
	lastVersion := h.versionService.GetLastVersionInfo()
	return &version.Version{
		ServerVersion: h.Version,
		FiasVersion:   strconv.Itoa(lastVersion.ID),
		GrpcVersion:   grpc.Version,
	}, nil
}
