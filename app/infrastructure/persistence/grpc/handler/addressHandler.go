package handler

import (
	"context"
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/service"
	addressV1 "github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/v1/address"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

type AddressHandler struct {
	Server         *grpc.Server
	addressService *service.AddressService
	houseService   *service.HouseService
}

func NewAddressHandler(a *service.AddressService, h *service.HouseService) *AddressHandler {
	handler := &AddressHandler{
		addressService: a,
		houseService:   h,
	}

	return handler
}

func (h *AddressHandler) GetCitiesByTerm(ctx context.Context, request *addressV1.TermRequest) (*addressV1.AddressListResponse, error) {
	if request.Term == "" {
		return nil, status.Error(codes.InvalidArgument, "term is required")
	}
	cities := h.addressService.GetCitiesByTerm(request.Term, request.Size, request.From)
	return h.prepareList(cities)
}

func (h *AddressHandler) GetAddressByTerm(ctx context.Context, request *addressV1.TermFilterRequest) (*addressV1.AddressListResponse, error) {
	if request.Term == "" {
		return nil, status.Error(codes.InvalidArgument, "term is required")
	}
	filters := h.prepareFilter(request.Filter)
	cities := h.addressService.GetAddressByTerm(request.Term, request.Size, request.From, filters...)
	return h.prepareList(cities)
}

func (h *AddressHandler) GetAddressByPostal(ctx context.Context, request *addressV1.TermRequest) (*addressV1.AddressListResponse, error) {
	if request.Term == "" {
		return nil, status.Error(codes.InvalidArgument, "term is required")
	}
	cities := h.addressService.GetAddressByPostal(request.Term, request.Size, request.From)
	return h.prepareList(cities)
}

func (h *AddressHandler) GetAllCities(ctx context.Context, empty *empty.Empty) (*addressV1.AddressListResponse, error) {
	cities := h.addressService.GetCities()
	return h.prepareList(cities)
}

func (h *AddressHandler) GetByGuid(ctx context.Context, guid *addressV1.GuidRequest) (*addressV1.Address, error) {
	if guid.Guid == "" {
		return nil, status.Error(codes.InvalidArgument, "guid is required")
	}
	addr := h.addressService.GetByGuid(guid.Guid)
	if addr != nil {
		return h.convertToAddress(addr), nil
	}

	return nil, status.Error(codes.NotFound, "address not found")
}

func (h *AddressHandler) GetSuggests(ctx context.Context, request *addressV1.SimpleTermFilterRequest) (*addressV1.AddressListResponse, error) {
	var houseNum int64
	size := request.Size
	if size == 0 {
		size = 100
	}
	filters := h.prepareFilter(request.Filter)

	suggests := h.addressService.GetAddressByTerm(request.Term, size, 0, filters...)
	houseNum = size - int64(len(suggests))
	if houseNum > 0 {
		cities := make(map[string]*entity.AddressObject, houseNum)
		houses := h.houseService.GetAddressByTerm(request.Term, houseNum, 0, filters...)
		for _, house := range houses {
			city, ok := cities[house.AoGuid]
			if ok == false {
				city = h.addressService.GetByGuid(house.AoGuid)
				if city == nil {
					continue
				}
				cities[house.AoGuid] = city
			}

			suggests = append(suggests, &entity.AddressObject{
				ID:             house.ID,
				AoGuid:         house.HouseGuid,
				ParentGuid:     house.AoGuid,
				FormalName:     house.HouseFullNum,
				ShortName:      "",
				AoLevel:        8,
				OffName:        house.HouseFullNum,
				Code:           city.Code,
				RegionCode:     city.RegionCode,
				PostalCode:     house.PostalCode,
				Okato:          house.Okato,
				Oktmo:          house.Oktmo,
				ActStatus:      city.ActStatus,
				LiveStatus:     city.LiveStatus,
				CurrStatus:     city.CurrStatus,
				StartDate:      house.StartDate,
				EndDate:        house.EndDate,
				UpdateDate:     house.UpdateDate,
				FullName:       house.HouseFullNum,
				FullAddress:    house.FullAddress,
				District:       city.District,
				DistrictType:   city.DistrictType,
				DistrictFull:   city.DistrictFull,
				Settlement:     city.Settlement,
				SettlementType: city.SettlementType,
				SettlementFull: city.SettlementFull,
				Street:         city.Street,
				StreetType:     city.StreetType,
				StreetFull:     city.StreetFull,
			})
		}
	}

	return h.prepareList(suggests)
}

func (h *AddressHandler) prepareFilter(requestFilter *addressV1.FilterObject) []entity.FilterObject {
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

func (h *AddressHandler) prepareList(cities []*entity.AddressObject) (*addressV1.AddressListResponse, error) {
	list := addressV1.AddressListResponse{}

	for _, city := range cities {
		list.Items = append(list.Items, h.convertToAddress(city))
	}

	return &list, nil
}

func (h *AddressHandler) convertToAddress(addr *entity.AddressObject) *addressV1.Address {
	if addr == nil {
		return nil
	}

	return &addressV1.Address{
		ID:             addr.ID,
		AoGuid:         addr.AoGuid,
		AoLevel:        strconv.Itoa(addr.AoLevel),
		FormalName:     addr.FormalName,
		ParentGuid:     addr.ParentGuid,
		ShortName:      addr.ShortName,
		PostalCode:     addr.PostalCode,
		FullName:       addr.FullName,
		FullAddress:    addr.FullAddress,
		District:       addr.District,
		DistrictType:   addr.DistrictType,
		DistrictFull:   addr.DistrictFull,
		Settlement:     addr.Settlement,
		SettlementType: addr.SettlementType,
		SettlementFull: addr.SettlementFull,
		Street:         addr.Street,
		StreetType:     addr.StreetType,
		StreetFull:     addr.StreetFull,
		KladrId:        addr.Code,
	}
}
