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

func (h *AddressHandler) GetCitiesByTerm(ctx context.Context, request *addressV1.TermRequest) (*addressV1.AddressListResponse, error) {
	cities := h.addressService.GetCitiesByTerm(request.Term, request.Size, request.From)
	return h.prepareList(cities)
}

func (h *AddressHandler) GetAddressByTerm(ctx context.Context, request *addressV1.TermRequest) (*addressV1.AddressListResponse, error) {
	cities := h.addressService.GetAddressByTerm(request.Term, request.Size, request.From)
	return h.prepareList(cities)
}

func (h *AddressHandler) GetAllCities(ctx context.Context, empty *empty.Empty) (*addressV1.AddressListResponse, error) {
	cities := h.addressService.GetCities()
	return h.prepareList(cities)
}

func (h *AddressHandler) GetByGuid(ctx context.Context, guid *addressV1.GuidRequest) (*addressV1.Address, error) {
	addr := h.addressService.GetByGuid(guid.Guid)
	if addr != nil {
		return h.convertToAddress(addr), nil
	}

	return nil, status.Error(codes.NotFound, "address not found")
}

func (h *AddressHandler) Serve() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	return h.Server.Serve(listener)
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
}
