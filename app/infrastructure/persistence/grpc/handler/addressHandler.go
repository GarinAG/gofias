package handler

import (
	"context"
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/service"
	fiasV1 "github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/v1/fias"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
)

// GRPC-обработчик адресов
type AddressHandler struct {
	Server         *grpc.Server            // GRPC-сервер
	addressService *service.AddressService // Сервис адресов
	houseService   *service.HouseService   // Сервис домов
}

// Инициализация обработчика
func NewAddressHandler(a *service.AddressService, h *service.HouseService) *AddressHandler {
	handler := &AddressHandler{
		addressService: a,
		houseService:   h,
	}

	return handler
}

// Найти города по подстроке
func (h *AddressHandler) GetCitiesByTerm(ctx context.Context, request *fiasV1.TermRequest) (*fiasV1.AddressListResponse, error) {
	if request.Term == "" {
		return nil, status.Error(codes.InvalidArgument, "term is required")
	}
	cities := h.addressService.GetCitiesByTerm(request.Term, request.Size, request.From)
	return h.prepareList(cities)
}

// Найти адрес по подстроке
func (h *AddressHandler) GetAddressByTerm(ctx context.Context, request *fiasV1.TermFilterRequest) (*fiasV1.AddressListResponse, error) {
	if request.Term == "" {
		return nil, status.Error(codes.InvalidArgument, "term is required")
	}
	filters := h.prepareFilter(request.Filter)
	cities := h.addressService.GetAddressByTerm(request.Term, request.Size, request.From, filters...)
	return h.prepareList(cities)
}

// Найти адрес по почтовому индексу
func (h *AddressHandler) GetAddressByPostal(ctx context.Context, request *fiasV1.TermRequest) (*fiasV1.AddressListResponse, error) {
	if request.Term == "" {
		return nil, status.Error(codes.InvalidArgument, "term is required")
	}
	cities := h.addressService.GetAddressByPostal(request.Term, request.Size, request.From)
	return h.prepareList(cities)
}

// Получить список всех городов
func (h *AddressHandler) GetAllCities(ctx context.Context, empty *empty.Empty) (*fiasV1.AddressListResponse, error) {
	cities := h.addressService.GetCities()
	return h.prepareList(cities)
}

// Найти адрес по GUID
func (h *AddressHandler) GetByGuid(ctx context.Context, guid *fiasV1.GuidRequest) (*fiasV1.Address, error) {
	if guid.Guid == "" {
		return nil, status.Error(codes.InvalidArgument, "guid is required")
	}
	addr := h.addressService.GetByGuid(guid.Guid)
	if addr != nil {
		return h.convertToAddress(addr), nil
	}

	return nil, status.Error(codes.NotFound, "address not found")
}

// Найти адрес по подстроке
func (h *AddressHandler) GetSuggests(ctx context.Context, request *fiasV1.SimpleTermFilterRequest) (*fiasV1.AddressListResponse, error) {
	if request.Term == "" {
		return nil, status.Error(codes.InvalidArgument, "term is required")
	}
	var houseNum int64
	size := request.Size
	// Ограничивает размер выборки
	if size == 0 {
		size = 100
	}
	filters := h.prepareFilter(request.Filter)

	// Получает адреса по подсроке
	suggests := h.addressService.GetAddressByTerm(request.Term, size, 0, filters...)
	houseNum = size - int64(len(suggests))
	// Проверка на необходимость загрузки домов
	if houseNum > 0 {
		cities := make(map[string]*entity.AddressObject, houseNum)
		// Получает дома по подсроке
		houses := h.houseService.GetAddressByTerm(request.Term, houseNum, 0, filters...)
		for _, house := range houses {
			// Ищет информацию об адресе дома в кэше
			city, ok := cities[house.AoGuid]
			if ok == false {
				// Получает информацию об адресе дома
				city = h.addressService.GetByGuid(house.AoGuid)
				if city == nil {
					continue
				}
				// Сохраняет информацию об адресе в кэш
				cities[house.AoGuid] = city
			}

			// Формирует объект адреса
			city.ID = house.ID
			city.AoGuid = house.HouseGuid
			city.ParentGuid = house.AoGuid
			city.FormalName = house.HouseFullNum
			city.ShortName = ""
			city.AoLevel = 8
			city.OffName = house.HouseFullNum
			city.PostalCode = house.PostalCode
			city.Okato = house.Okato
			city.Oktmo = house.Oktmo
			city.StartDate = house.StartDate
			city.EndDate = house.EndDate
			city.UpdateDate = house.UpdateDate
			city.FullName = house.HouseFullNum
			city.FullAddress = house.FullAddress
			city.BazisUpdateDate = house.BazisUpdateDate

			suggests = append(suggests, city)
		}
	}

	return h.prepareList(suggests)
}

// Подготавливает фильтр запросов
func (h *AddressHandler) prepareFilter(requestFilter *fiasV1.FilterObject) []entity.FilterObject {
	filter := entity.FilterObject{}
	if requestFilter != nil {
		if requestFilter.Level != nil {
			filter.Level = entity.NumberFilter{
				Values: requestFilter.Level.Values,
				Min:    requestFilter.Level.Min,
				Max:    requestFilter.Level.Max,
			}
		}
		if requestFilter.ParentGuid != nil {
			filter.ParentGuid = entity.StringFilter{
				Values: requestFilter.ParentGuid.Values,
			}
		}
		if requestFilter.KladrId != nil {
			filter.KladrId = entity.StringFilter{
				Values: requestFilter.KladrId.Values,
			}
		}
	}

	return []entity.FilterObject{
		filter,
	}
}

// Формирует список объектов адресов
func (h *AddressHandler) prepareList(cities []*entity.AddressObject) (*fiasV1.AddressListResponse, error) {
	list := fiasV1.AddressListResponse{}

	for _, city := range cities {
		list.Items = append(list.Items, h.convertToAddress(city))
	}

	return &list, nil
}

// Конвертирует объект адреса в grpc-объект
func (h *AddressHandler) convertToAddress(addr *entity.AddressObject) *fiasV1.Address {
	if addr == nil {
		return nil
	}

	item := fiasV1.Address{
		ID:                addr.ID,
		FiasId:            addr.AoGuid,
		FiasLevel:         strconv.Itoa(addr.AoLevel),
		ParentFiasId:      addr.ParentGuid,
		ShortName:         addr.ShortName,
		FormalName:        addr.FormalName,
		PostalCode:        addr.PostalCode,
		FullName:          addr.FullName,
		FullAddress:       addr.FullAddress,
		KladrId:           addr.Code,
		RegionFiasId:      addr.RegionGuid,
		RegionKladrId:     addr.RegionKladr,
		Region:            addr.Region,
		RegionType:        addr.RegionType,
		RegionFull:        addr.RegionFull,
		AreaFiasId:        addr.AreaGuid,
		AreaKladrId:       addr.AreaKladr,
		Area:              addr.Area,
		AreaType:          addr.AreaType,
		AreaFull:          addr.AreaFull,
		CityFiasId:        addr.CityGuid,
		CityKladrId:       addr.CityKladr,
		City:              addr.City,
		CityType:          addr.CityType,
		CityFull:          addr.CityFull,
		SettlementFiasId:  addr.SettlementGuid,
		SettlementKladrId: addr.SettlementKladr,
		Settlement:        addr.Settlement,
		SettlementType:    addr.SettlementType,
		SettlementFull:    addr.SettlementFull,
		StreetFiasId:      addr.StreetGuid,
		StreetKladrId:     addr.StreetKladr,
		Street:            addr.Street,
		StreetType:        addr.StreetType,
		StreetFull:        addr.StreetFull,
		Okato:             addr.Okato,
		Oktmo:             addr.Oktmo,
		UpdatedDate:       addr.BazisUpdateDate,
	}

	if addr.AoLevel == 8 {
		item.HouseFiasId = addr.AoGuid
		item.HouseKladrId = addr.Code
		item.House = addr.FormalName
		item.HouseType = addr.ShortName
		item.HouseFull = addr.FullName
		item.StreetFiasId = addr.ParentGuid
	} else if addr.AoLevel == 7 {
		item.StreetFiasId = addr.AoGuid
		item.StreetKladrId = addr.Code
		item.Street = addr.FormalName
		item.StreetType = addr.ShortName
		item.StreetFull = addr.FullName
	} else if addr.AoLevel == 5 || addr.AoLevel == 6 {
		item.SettlementFiasId = addr.AoGuid
		item.SettlementKladrId = addr.Code
		item.Settlement = addr.FormalName
		item.SettlementType = addr.ShortName
		item.SettlementFull = addr.FullName
	} else if addr.AoLevel == 4 {
		item.CityFiasId = addr.AoGuid
		item.CityKladrId = addr.Code
		item.City = addr.FormalName
		item.CityType = addr.ShortName
		item.CityFull = addr.FullName
	} else if addr.AoLevel == 3 {
		item.AreaFiasId = addr.AoGuid
		item.AreaKladrId = addr.Code
		item.Area = addr.FormalName
		item.AreaType = addr.ShortName
		item.AreaFull = addr.FullName
	} else if addr.AoLevel <= 2 {
		item.RegionFiasId = addr.AoGuid
		item.RegionKladrId = addr.Code
		item.Region = addr.FormalName
		item.RegionType = addr.ShortName
		item.RegionFull = addr.FullName
	}
	if addr.Location != "" {
		location := strings.Split(addr.Location, ",")
		if len(location) == 2 {
			lat, err := strconv.ParseFloat(location[0], 32)
			if err == nil {
				item.GeoLat = float32(lat)
			}
			lon, err := strconv.ParseFloat(location[1], 32)
			if err == nil {
				item.GeoLon = float32(lon)
			}
		}
	}

	return &item
}
