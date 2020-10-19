package handler

import (
	"context"
	"github.com/GarinAG/gofias/domain/version/service"
	"github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/version"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"strconv"
)

// GRPC-обработчик проверки версии сервиса
type VersionHandler struct {
	Version        string                  // Версия приложения
	Server         *grpc.Server            // GRPC-сервер
	versionService *service.VersionService // Сервис версий
}

// Инициализация обработчика
func NewVersionHandler(a *service.VersionService, v string) *VersionHandler {
	handler := &VersionHandler{
		Version:        v,
		versionService: a,
	}

	return handler
}

// Получить информацию о версии приложения
func (h *VersionHandler) GetVersion(ctx context.Context, empty *empty.Empty) (*version.Version, error) {
	// Получает поледнюю версию БД ФИАС
	lastVersion := h.versionService.GetLastVersionInfo()
	return &version.Version{
		ServerVersion: h.Version,
		FiasVersion:   strconv.Itoa(lastVersion.ID),
		GrpcVersion:   grpc.Version,
	}, nil
}
