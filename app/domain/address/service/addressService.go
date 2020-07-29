package service

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/interfaces"
)

type AddressService struct {
	addressRepo repository.AddressRepositoryInterface
	logger      interfaces.LoggerInterface
}

func NewAddressService(addressRepo repository.AddressRepositoryInterface, logger interfaces.LoggerInterface) *AddressService {
	return &AddressService{
		addressRepo: addressRepo,
		logger:      logger,
	}
}

func (a *AddressService) GetByGuid(guid string) *entity.AddressObject {
	address, err := a.addressRepo.GetByGuid(guid)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return address
}

func (a *AddressService) GetCities() []*entity.AddressObject {
	cities, err := a.addressRepo.GetCities()
	if err != nil {
		a.logger.Error(err.Error())
	}

	return cities
}

func (a *AddressService) GetCitiesByTerm(term string, count int64) []*entity.AddressObject {
	cities, err := a.addressRepo.GetCitiesByTerm(term, count)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return cities
}
