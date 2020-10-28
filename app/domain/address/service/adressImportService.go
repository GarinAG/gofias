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

// Сервис импорта адресов
type AddressImportService struct {
	AddressRepo repository.AddressRepositoryInterface // Репозиторий адресов
	IsFull      bool                                  `default:"false"` // Полный импорт
	logger      interfaces.LoggerInterface            // Логгер
}

// Инициализация сервиса
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

// Импорт адресов
func (a *AddressImportService) Import(filePath string, wg *sync.WaitGroup, cnt chan int) {
	defer wg.Done()
	var importWg sync.WaitGroup
	importWg.Add(2)
	addressChannel := make(chan interface{})
	// Чтение файла импорта и парсинг элементов
	go util.ParseFile(&importWg, filePath, addressChannel, a.logger, a.ParseElement, "Object", -1)
	// Сохраняет элементы в БД
	go a.AddressRepo.InsertUpdateCollection(&importWg, addressChannel, cnt, a.IsFull)
	importWg.Wait()
}

// Индексация таблицы адресов
func (a *AddressImportService) Index(isFull bool, start time.Time, guids []string, wg *sync.WaitGroup, indexChan chan<- entity.IndexObject) {
	defer wg.Done()
	err := a.AddressRepo.Index(isFull, start, guids, indexChan)
	a.checkError(err)
}

// Разбор объекта из xml
func (a *AddressImportService) ParseElement(element *xmlparser.XMLElement) (interface{}, error) {
	// Пропускает неактивные элементы при полном импорте
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

// Подсчет общего количества адресов
func (a *AddressImportService) CountAllData() int64 {
	res, err := a.AddressRepo.CountAllData(nil)
	a.checkError(err)

	return res
}

// Проверяет наличие ошибки и логирует ее
func (a *AddressImportService) checkError(err error) {
	if err != nil {
		a.logger.Error(err.Error())
	}
}

// Получить список адресов по GUID
func (a *AddressImportService) GetAddressByGuidList(guids []string) ([]*entity.AddressObject, error) {
	res, err := a.AddressRepo.GetAddressByGuidList(util.UniqueStringSlice(guids))
	a.checkError(err)

	return res, nil
}
