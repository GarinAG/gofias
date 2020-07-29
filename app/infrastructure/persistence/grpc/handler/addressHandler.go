package handler

import (
	"context"
	"github.com/GarinAG/gofias/domain/address/service"
	address_v1 "github.com/GarinAG/gofias/interfaces/grpc/proto/v1/address"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"net"
	"strconv"
)

type AddressHandler struct {
	Server         *grpc.Server
	addressService *service.AddressService
}

func NewAddressHandler(a *service.AddressService) *AddressHandler {
	handler := &AddressHandler{
		addressService: a,
	}

	return handler
}

func (h *AddressHandler) GetCitiesByTerm(request *address_v1.TermRequest, stream address_v1.AddressHandler_GetCitiesByTermServer) error {
	if request.Count == 0 {
		request.Count = 10
	}

	cities := h.addressService.GetCitiesByTerm(request.Term, request.Count)

	for _, city := range cities {
		stream.Send(&address_v1.Address{
			AoGuid:         city.AoGuid,
			AoLevel:        strconv.Itoa(city.AoLevel),
			FormalName:     city.FormalName,
			ParentGuid:     city.ParentGuid,
			ShortName:      city.ShortName,
			PostalCode:     city.PostalCode,
			FullName:       city.FullName,
			FullAddress:    city.FullAddress,
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

	return nil
}

func (h *AddressHandler) GetAllCities(empty *empty.Empty, stream address_v1.AddressHandler_GetAllCitiesServer) error {
	cities := h.addressService.GetCities()

	for _, city := range cities {
		stream.Send(&address_v1.Address{
			AoGuid:         city.AoGuid,
			AoLevel:        strconv.Itoa(city.AoLevel),
			FormalName:     city.FormalName,
			ParentGuid:     city.ParentGuid,
			ShortName:      city.ShortName,
			PostalCode:     city.PostalCode,
			FullName:       city.FullName,
			FullAddress:    city.FullAddress,
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

	return nil
}

func (h *AddressHandler) GetByGuid(ctx context.Context, guid *address_v1.GuidRequest) (*address_v1.Address, error) {
	addr := h.addressService.GetByGuid(guid.Guid)

	result := address_v1.Address{
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
	}

	return &result, nil
}

func (h *AddressHandler) Serve() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	return h.Server.Serve(listener)
}
