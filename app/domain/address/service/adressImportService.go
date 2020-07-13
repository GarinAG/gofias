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

type AddressImportService struct {
	AddressRepo repository.AddressRepositoryInterface
	logger      interfaces.LoggerInterface
}

func NewAddressService(addressRepo repository.AddressRepositoryInterface, logger interfaces.LoggerInterface) *AddressImportService {
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
	go util.ParseFile(filePath, done, addressChannel, a.logger, a.ParseElement, "Object")
	go a.AddressRepo.InsertUpdateCollection(addressChannel, done, cnt)
}

func (a *AddressImportService) Index(isFull bool, start time.Time, housesCount int64, GetHousesByGuid repository.GetHousesByGuid) {
	err := a.AddressRepo.Index(isFull, start, housesCount, GetHousesByGuid)
	if err != nil {
		a.logger.Error(err.Error())
	}
}

func (a *AddressImportService) ParseElement(element *xmlparser.XMLElement) (interface{}, error) {
	result := entity.AddressObject{
		ID:         element.Attrs["AOID"],
		AoGuid:     element.Attrs["AOGUID"],
		ParentGuid: element.Attrs["PARENTGUID"],
		FormalName: element.Attrs["FORMALNAME"],
		ShortName:  element.Attrs["SHORTNAME"],
		AoLevel:    element.Attrs["AOLEVEL"],
		OffName:    element.Attrs["OFFNAME"],
		AreaCode:   element.Attrs["AREACODE"],
		CityCode:   element.Attrs["CITYCODE"],
		PlaceCode:  element.Attrs["PLACECODE"],
		AutoCode:   element.Attrs["AUTOCODE"],
		PlanCode:   element.Attrs["PLANCODE"],
		StreetCode: element.Attrs["STREETCODE"],
		CTarCode:   element.Attrs["CTARCODE"],
		ExtrCode:   element.Attrs["EXTRCODE"],
		SextCode:   element.Attrs["SEXTCODE"],
		Code:       element.Attrs["CODE"],
		RegionCode: element.Attrs["REGIONCODE"],
		PlainCode:  element.Attrs["PLAINCODE"],
		PostalCode: element.Attrs["POSTALCODE"],
		Okato:      element.Attrs["OKATO"],
		Oktmo:      element.Attrs["OKTMO"],
		IfNsFl:     element.Attrs["IFNSFL"],
		IfNsUl:     element.Attrs["IFNSUL"],
		TerrIfNsFl: element.Attrs["TERRIFNSFL"],
		TerrIfNsUl: element.Attrs["TERRIFNSUL"],
		NormDoc:    element.Attrs["NORMDOC"],
		ActStatus:  element.Attrs["ACTSTATUS"],
		LiveStatus: element.Attrs["LIVESTATUS"],
		CurrStatus: element.Attrs["CURRSTATUS"],
		OperStatus: element.Attrs["OPERSTATUS"],
		StartDate:  element.Attrs["STARTDATE"],
		EndDate:    element.Attrs["ENDDATE"],
		UpdateDate: element.Attrs["UPDATEDATE"],
	}

	return result, nil
}

func (a *AddressImportService) GetByFormalName(term string) *entity.AddressObject {
	res, err := a.AddressRepo.GetByFormalName(term)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return res
}
func (a *AddressImportService) GetByGuid(guid string) *entity.AddressObject {
	res, err := a.AddressRepo.GetByGuid(guid)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return res
}
func (a *AddressImportService) GetCities() []*entity.AddressObject {
	res, err := a.AddressRepo.GetCities()
	if err != nil {
		a.logger.Error(err.Error())
	}

	return res
}

func (a *AddressImportService) GetCitiesByTerm(term string, count int64) []*entity.AddressObject {
	res, err := a.AddressRepo.GetCitiesByTerm(term, count)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return res
}

func (a *AddressImportService) GetCityByFormalName(term string) *entity.AddressObject {
	res, err := a.AddressRepo.GetCityByFormalName(term)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return res
}

func (a *AddressImportService) CountAllData() int64 {
	res, err := a.AddressRepo.CountAllData()
	if err != nil {
		a.logger.Error(err.Error())
	}

	return res
}
