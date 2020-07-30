package handler

import (
	"context"
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/service"
	addressV1 "github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/v1/address"
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

func (h *AddressHandler) GetCitiesByTerm(request *addressV1.TermRequest, stream addressV1.AddressHandler_GetCitiesByTermServer) error {
	if request.Count == 0 {
		request.Count = 10
	}

	cities := h.addressService.GetCitiesByTerm(request.Term, request.Count, request.Size, request.From)

	for _, city := range cities {
		stream.Send(h.convertToAddress(city))
	}

	return nil
}

func (h *AddressHandler) GetAddressByTerm(request *addressV1.TermRequest, stream addressV1.AddressHandler_GetAddressByTermServer) error {
	if request.Count == 0 {
		request.Count = 10
	}

	cities := h.addressService.GetAddressByTerm(request.Term, request.Count, request.Size, request.From)

	for _, city := range cities {
		stream.Send(h.convertToAddress(city))
	}

	return nil
}

func (h *AddressHandler) GetAllCities(empty *empty.Empty, stream addressV1.AddressHandler_GetAllCitiesServer) error {
	cities := h.addressService.GetCities()

	for _, city := range cities {
		stream.Send(h.convertToAddress(city))
	}

	return nil
}

func (h *AddressHandler) GetByGuid(ctx context.Context, guid *addressV1.GuidRequest) (*addressV1.Address, error) {
	addr := h.addressService.GetByGuid(guid.Guid)
	result := h.convertToAddress(addr)

	return result, nil
}

func (h *AddressHandler) Serve() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	return h.Server.Serve(listener)
}

func (h *AddressHandler) convertToAddress(addr *entity.AddressObject) *addressV1.Address {
	if addr == nil {
		return nil
	}

	result := addressV1.Address{
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

	return &result
}
