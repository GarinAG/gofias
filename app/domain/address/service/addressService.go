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

func (a *AddressService) GetCitiesByTerm(term string, size int64, from int64) []*entity.AddressObject {
	cities, err := a.addressRepo.GetCitiesByTerm(term, size, from)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return cities
}

func (a *AddressService) GetAddressByTerm(term string, size int64, from int64, filter ...entity.FilterObject) []*entity.AddressObject {
	cities, err := a.addressRepo.GetAddressByTerm(term, size, from, filter...)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return cities
}

func (a *AddressService) GetAddressByPostal(term string, size int64, from int64) []*entity.AddressObject {
	cities, err := a.addressRepo.GetAddressByPostal(term, size, from)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return cities
}
