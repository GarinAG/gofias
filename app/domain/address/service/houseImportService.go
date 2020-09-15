package service

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	xmlparser "github.com/tamerh/xml-stream-parser"
	"os"
	"sync"
	"time"
)

type HouseImportService struct {
	HouseRepo   repository.HouseRepositoryInterface
	IsFull      bool `default:"false"`
	logger      interfaces.LoggerInterface
	currentTime int64
}

func NewHouseImportService(houseRepo repository.HouseRepositoryInterface, logger interfaces.LoggerInterface) *HouseImportService {
	err := houseRepo.Init()
	if err != nil {
		logger.Panic(err.Error())
		os.Exit(1)
	}

	return &HouseImportService{
		HouseRepo:   houseRepo,
		logger:      logger,
		currentTime: time.Now().Unix(),
	}
}

func (h *HouseImportService) GetRepo() repository.HouseRepositoryInterface {
	return h.HouseRepo
}

func (h *HouseImportService) Import(filePath string, wg *sync.WaitGroup, cnt chan int) {
	defer wg.Done()
	addressChannel := make(chan interface{})
	done := make(chan bool)
	total := 0
	if h.IsFull {
		total = 75000000
	}
	go util.ParseFile(filePath, done, addressChannel, h.logger, h.ParseElement, "House", total)
	go h.HouseRepo.InsertUpdateCollection(addressChannel, done, cnt, h.IsFull)
}

func (h *HouseImportService) ParseElement(element *xmlparser.XMLElement) (interface{}, error) {
	if h.IsFull {
		end, err := time.Parse("2006-01-02", element.Attrs["ENDDATE"])

		if err != nil || end.Unix() <= h.currentTime {
			return nil, nil
		}
	}

	result := entity.HouseObject{
		ID:         element.Attrs["HOUSEID"],
		AoGuid:     element.Attrs["AOGUID"],
		HouseNum:   element.Attrs["HOUSENUM"],
		PostalCode: element.Attrs["POSTALCODE"],
		Okato:      element.Attrs["OKATO"],
		Oktmo:      element.Attrs["OKTMO"],
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

func (h *HouseImportService) GetLastUpdatedGuids(start time.Time) []string {
	res, err := h.HouseRepo.GetLastUpdatedGuids(start)
	if err != nil {
		h.logger.Error(err.Error())
	}

	return res
}

func (h *HouseImportService) CountAllData() int64 {
	res, err := h.HouseRepo.CountAllData(nil)
	if err != nil {
		h.logger.Error(err.Error())
	}

	return res
}
