package service

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/interfaces"
	"os"
)

// Сервис получения данных о домах
type HouseService struct {
	HouseRepo repository.HouseRepositoryInterface // Репозиторий домов
	logger    interfaces.LoggerInterface          // Логгер
}

// Инициализация сервиса
func NewHouseService(houseRepo repository.HouseRepositoryInterface, logger interfaces.LoggerInterface) *HouseService {
	err := houseRepo.Init()
	if err != nil {
		logger.Panic(err.Error())
		os.Exit(1)
	}

	return &HouseService{
		HouseRepo: houseRepo,
		logger:    logger,
	}
}

// Найти дома по GUID адреса
func (h *HouseService) GetByAddressGuid(giud string) []*entity.HouseObject {
	res, err := h.HouseRepo.GetByAddressGuid(giud)
	h.checkError(err)

	return res
}

// Найти дома по подстроке
func (h *HouseService) GetAddressByTerm(term string, size int64, from int64) []*entity.HouseObject {
	houses, err := h.HouseRepo.GetAddressByTerm(term, size, from)
	h.checkError(err)

	return houses
}

// Проверяет наличие ошибки и логирует ее
func (h *HouseService) checkError(err error) {
	if err != nil {
		h.logger.Error(err.Error())
	}
}
