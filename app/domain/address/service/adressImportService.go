package service

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	xmlparser "github.com/tamerh/xml-stream-parser"
	"os"
	"strconv"
	"sync"
	"time"
)

type AddressImportService struct {
	AddressRepo repository.AddressRepositoryInterface
	IsFull      bool `default:"false"`
	logger      interfaces.LoggerInterface
}

func NewAddressImportService(addressRepo repository.AddressRepositoryInterface, logger interfaces.LoggerInterface) *AddressImportService {
	err := addressRepo.Init()
	if err != nil {
		logger.Panic(err.Error())
		os.Exit(1)
	}

	return &AddressImportService{
		AddressRepo: addressRepo,
		logger:      logger,
	}
}

func (a *AddressImportService) Import(filePath string, wg *sync.WaitGroup, cnt chan int) {
	defer wg.Done()
	addressChannel := make(chan interface{})
	done := make(chan bool)
	total := 0
	if a.IsFull {
		total = 4500000
	}
	go util.ParseFile(filePath, done, addressChannel, a.logger, a.ParseElement, "Object", total)
	go a.AddressRepo.InsertUpdateCollection(addressChannel, done, cnt, a.IsFull)
}

func (a *AddressImportService) Index(isFull bool, start time.Time, guids []string) {
	err := a.AddressRepo.Index(isFull, start, guids)
	if err != nil {
		a.logger.Error(err.Error())
	}
}

func (a *AddressImportService) ParseElement(element *xmlparser.XMLElement) (interface{}, error) {
	if a.IsFull {
		if element.Attrs["CURRSTATUS"] != "0" ||
			element.Attrs["ACTSTATUS"] != "1" ||
			element.Attrs["LIVESTATUS"] != "1" {

			return nil, nil
		}
	}
	level, _ := strconv.Atoi(element.Attrs["AOLEVEL"])

	result := entity.AddressObject{
		ID:         element.Attrs["AOID"],
		AoGuid:     element.Attrs["AOGUID"],
		ParentGuid: element.Attrs["PARENTGUID"],
		FormalName: element.Attrs["FORMALNAME"],
		ShortName:  element.Attrs["SHORTNAME"],
		AoLevel:    level,
		OffName:    element.Attrs["OFFNAME"],
		Code:       element.Attrs["CODE"],
		RegionCode: element.Attrs["REGIONCODE"],
		PostalCode: element.Attrs["POSTALCODE"],
		Okato:      element.Attrs["OKATO"],
		Oktmo:      element.Attrs["OKTMO"],
		ActStatus:  element.Attrs["ACTSTATUS"],
		LiveStatus: element.Attrs["LIVESTATUS"],
		CurrStatus: element.Attrs["CURRSTATUS"],
		StartDate:  element.Attrs["STARTDATE"],
		EndDate:    element.Attrs["ENDDATE"],
		UpdateDate: element.Attrs["UPDATEDATE"],
	}

	return result, nil
}

func (a *AddressImportService) CountAllData() int64 {
	res, err := a.AddressRepo.CountAllData(nil)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return res
}
