package service

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/interfaces"
)

// Сервис получения данных об адресах
type AddressService struct {
	addressRepo repository.AddressRepositoryInterface // Репозиторий адресов
	logger      interfaces.LoggerInterface            // Логгер
}

// Инициализация сервиса
func NewAddressService(addressRepo repository.AddressRepositoryInterface, logger interfaces.LoggerInterface) *AddressService {
	return &AddressService{
		addressRepo: addressRepo,
		logger:      logger,
	}
}

// Найти адрес по GUID
func (a *AddressService) GetByGuid(guid string) *entity.AddressObject {
	address, err := a.addressRepo.GetByGuid(guid)
	a.checkError(err)

	return address
}

// Получить список всех городов
func (a *AddressService) GetCities() []*entity.AddressObject {
	cities, err := a.addressRepo.GetCities()
	a.checkError(err)

	return cities
}

// Найти города по подстроке
func (a *AddressService) GetCitiesByTerm(term string, size int64, from int64) []*entity.AddressObject {
	cities, err := a.addressRepo.GetCitiesByTerm(term, size, from)
	a.checkError(err)

	return cities
}

// Найти адрес по подстроке
func (a *AddressService) GetAddressByTerm(term string, size int64, from int64, filter ...entity.FilterObject) []*entity.AddressObject {
	cities, err := a.addressRepo.GetAddressByTerm(term, size, from, filter...)
	a.checkError(err)

	return cities
}

// Найти адрес по почтовому индексу
func (a *AddressService) GetAddressByPostal(term string, size int64, from int64) []*entity.AddressObject {
	cities, err := a.addressRepo.GetAddressByPostal(term, size, from)
	a.checkError(err)

	return cities
}

// Проверяет наличие ошибки и логирует ее
func (a *AddressService) checkError(err error) {
	if err != nil {
		a.logger.Error(err.Error())
	}
}
