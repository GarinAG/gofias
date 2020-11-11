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

// Сервис импорта домов
type HouseImportService struct {
	HouseRepo   repository.HouseRepositoryInterface // Репозиторий домов
	IsFull      bool                                `default:"false"` // Полный импорт
	logger      interfaces.LoggerInterface          // Логгер
	currentTime int64                               // Время начала импорта
}

// Инициализация сервиса
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

// Импорт домов
func (h *HouseImportService) Import(filePath string, wg *sync.WaitGroup, cnt chan int) {
	defer wg.Done()
	var importWg sync.WaitGroup
	importWg.Add(2)
	addressChannel := make(chan interface{})
	// Чтение файла импорта и парсинг элементов
	go util.ParseFile(&importWg, filePath, addressChannel, h.logger, h.ParseElement, "House", -1)
	// Сохраняет элементы в БД
	go h.HouseRepo.InsertUpdateCollection(&importWg, addressChannel, cnt, h.IsFull)
	importWg.Wait()
}

// Разбор объекта из xml
func (h *HouseImportService) ParseElement(element *xmlparser.XMLElement) (interface{}, error) {
	// Пропускает неактивные элементы при полном импорте
	if h.IsFull {
		end, err := time.Parse("2006-01-02", element.Attrs["ENDDATE"])

		if err != nil || end.Unix() <= h.currentTime {
			return nil, nil
		}
	}

	result := entity.HouseObject{
		ID:         element.Attrs["HOUSEID"],
		AoGuid:     element.Attrs["AOGUID"],
		HouseGuid:  element.Attrs["HOUSEGUID"],
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

// Найти дома по GUID адреса
func (h *HouseImportService) GetByAddressGuid(giud string) []*entity.HouseObject {
	res, err := h.HouseRepo.GetByAddressGuid(giud)
	h.checkError(err)

	return res
}

// Получить последние обновленные дома
func (h *HouseImportService) GetLastUpdatedGuids(start time.Time) []string {
	res, err := h.HouseRepo.GetLastUpdatedGuids(start)
	h.checkError(err)

	return res
}

// Подсчитать общее количество домов в БД
func (h *HouseImportService) CountAllData() int64 {
	res, err := h.HouseRepo.CountAllData(nil)
	h.checkError(err)

	return res
}

// Индексация таблицы домов
func (h *HouseImportService) Index(start time.Time, wg *sync.WaitGroup, indexChan <-chan entity.IndexObject, objects repository.GetIndexObjects) {
	defer wg.Done()
	err := h.HouseRepo.Index(start, indexChan, objects)
	h.checkError(err)
}

// Проверяет наличие ошибки и логирует ее
func (h *HouseImportService) checkError(err error) {
	if err != nil {
		h.logger.Error(err.Error())
	}
}
