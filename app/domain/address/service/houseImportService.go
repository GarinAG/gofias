package service

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	xmlparser "github.com/tamerh/xml-stream-parser"
	"os"
	"sync"
)

type HouseImportService struct {
	HouseRepo repository.HouseRepositoryInterface
	logger    interfaces.LoggerInterface
}

func NewHouseService(houseRepo repository.HouseRepositoryInterface, logger interfaces.LoggerInterface) *HouseImportService {
	err := houseRepo.Init()
	if err != nil {
		logger.Panic(err.Error())
		os.Exit(1)
	}

	return &HouseImportService{
		HouseRepo: houseRepo,
		logger:    logger,
	}
}

func (h *HouseImportService) GetRepo() repository.HouseRepositoryInterface {
	return h.HouseRepo
}

func (h *HouseImportService) Import(filePath string, wg *sync.WaitGroup, cnt chan int) {
	defer wg.Done()
	addressChannel := make(chan interface{})
	done := make(chan bool)
	go util.ParseFile(filePath, done, addressChannel, h.logger, h.ParseElement, "House")
	go h.HouseRepo.InsertUpdateCollection(addressChannel, done, cnt)
}

func (h *HouseImportService) ParseElement(element *xmlparser.XMLElement) (interface{}, error) {
	result := entity.HouseObject{
		ID:         element.Attrs["HOUSEID"],
		AoGuid:     element.Attrs["AOGUID"],
		HouseNum:   element.Attrs["HOUSENUM"],
		RegionCode: element.Attrs["REGIONCODE"],
		PostalCode: element.Attrs["POSTALCODE"],
		Okato:      element.Attrs["OKATO"],
		Oktmo:      element.Attrs["OKTMO"],
		IfNsFl:     element.Attrs["IFNSFL"],
		IfNsUl:     element.Attrs["IFNSUL"],
		TerrIfNsFl: element.Attrs["TERRIFNSFL"],
		TerrIfNsUl: element.Attrs["TERRIFNSUL"],
		NormDoc:    element.Attrs["NORMDOC"],
		StartDate:  element.Attrs["STARTDATE"],
		EndDate:    element.Attrs["ENDDATE"],
		UpdateDate: element.Attrs["UPDATEDATE"],
		DivType:    element.Attrs["DIVTYPE"],
		BuildNum:   element.Attrs["BUILDNUM"],
		StructNum:  element.Attrs["STRUCNUM"],
		Counter:    element.Attrs["COUNTER"],
		CadNum:     element.Attrs["CADNUM"],
	}
	return result, nil
}

func (h *HouseImportService) GetByAddressGuid(giud string) []*entity.HouseObject {
	res, err := h.HouseRepo.GetByAddressGuid(giud)
	if err != nil {
		h.logger.Error(err.Error())
	}

	return res
}

func (h *HouseImportService) CountAllData() int64 {
	res, err := h.HouseRepo.CountAllData()
	if err != nil {
		h.logger.Error(err.Error())
	}

	return res
}
