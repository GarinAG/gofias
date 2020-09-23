package service

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/interfaces"
	"os"
)

type HouseService struct {
	HouseRepo repository.HouseRepositoryInterface
	logger    interfaces.LoggerInterface
}

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

func (h *HouseService) GetRepo() repository.HouseRepositoryInterface {
	return h.HouseRepo
}

func (h *HouseService) GetByAddressGuid(giud string) []*entity.HouseObject {
	res, err := h.HouseRepo.GetByAddressGuid(giud)
	if err != nil {
		h.logger.Error(err.Error())
	}

	return res
}

func (h *HouseService) GetAddressByTerm(term string, size int64, from int64) []*entity.HouseObject {
	houses, err := h.HouseRepo.GetAddressByTerm(term, size, from)
	if err != nil {
		h.logger.Error(err.Error())
	}

	return houses
}
