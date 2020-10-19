package service

import (
	addressEntity "github.com/GarinAG/gofias/domain/address/entity"
	directoryEntity "github.com/GarinAG/gofias/domain/directory/entity"
	"github.com/GarinAG/gofias/domain/directory/service"
	"github.com/GarinAG/gofias/domain/fiasApi/entity"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	versionEntity "github.com/GarinAG/gofias/domain/version/entity"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	"github.com/GarinAG/gofias/interfaces"
	"os"
	"regexp"
	"sync"
	"time"
)

// Общий сервис импорта
type ImportService struct {
	addressImportService *AddressImportService      // Сервис импорта адресов
	houseImportService   *HouseImportService        // Сервис импорта домов
	logger               interfaces.LoggerInterface // Логгер
	directoryService     *service.DirectoryService  // Сервис работы с файлами
	config               interfaces.ConfigInterface // Конфигурация
	IsFull               bool                       `default:"false"` // Полный импорт
	SkipHouses           bool                       `default:"false"` // Пропускать импорт домов
	SkipClear            bool                       `default:"false"` // Не удалять скачанные файлы после импорта
	Begin                time.Time                  // Время начала импорта
}

// Инициализация сервиса
func NewImportService(logger interfaces.LoggerInterface, ds *service.DirectoryService, addressImportService *AddressImportService, houseImportService *HouseImportService, config interfaces.ConfigInterface) *ImportService {
	return &ImportService{
		addressImportService: addressImportService,
		houseImportService:   houseImportService,
		logger:               logger,
		directoryService:     ds,
		config:               config,
		IsFull:               false,
		Begin:                time.Now(),
	}
}

// Получить список названий файлов импорта
func (is *ImportService) getParts() []string {
	parts := []string{addressEntity.AddressObject{}.GetXmlFile()}
	if !is.SkipHouses {
		parts = append(parts, addressEntity.HouseObject{}.GetXmlFile())
	}

	return parts
}

// Загрузка дельт
func (is *ImportService) StartDeltaImport(api *fiasApiService.FiasApiService, versionService *versionService.VersionService, version *versionEntity.Version) {
	// Получение полного списка версий ФИАС
	result := api.GetAllDownloadFileInfo()
	var needVersionList []entity.DownloadFileInfo
	// Проверка необходимости загрузки версии
	for _, file := range result {
		if file.VersionId == version.ID {
			break
		}
		needVersionList = append(needVersionList, file)
	}
	// Получает список названий файлов импорта
	parts := is.getParts()

	// Очищает директорию от ранее скачанных файлов
	is.clearDirectory(false)

	// Завершаем импорт, если скачана последняя версия
	if len(needVersionList) == 0 {
		is.logger.Info("Last version is uploaded")
		os.Exit(1)
	}
	// Идем от более ранней версии к более новой
	for i := len(needVersionList) - 1; i >= 0; i-- {
		uploadedVersion := needVersionList[i]
		cntAddr := 0
		cntHouses := 0

		is.logger.WithFields(interfaces.LoggerFields{
			"version": uploadedVersion,
		}).Debug("Uploaded version info")

		// Проверяет, есть ли ссылка на файл дельты
		if uploadedVersion.FiasDeltaXmlUrl != "" {
			// Загружает файл и распаковывает
			xmlFiles := is.directoryService.DownloadAndExtractFile(uploadedVersion.FiasDeltaXmlUrl, "fias_delta_xml.zip", parts...)
			// Читает xml-файлы и импортирует элементы
			cntAddr, cntHouses = is.ParseFiles(xmlFiles)
		}
		// Очищает директорию от ранее скачанных файлов
		is.clearDirectory(true)
		// Обновляет версию ФИАС в БД
		versionService.UpdateVersion(is.convertDownloadInfoToVersion(uploadedVersion, cntAddr, cntHouses))
	}

	is.logger.Info("Import finished")
}

// Загрузка полного импорта
func (is *ImportService) StartFullImport(api *fiasApiService.FiasApiService, versionService *versionService.VersionService) {
	is.IsFull = true
	is.addressImportService.IsFull = true
	is.houseImportService.IsFull = true

	// Получает ифнормацию о последней доступной версии ФИАС
	fileResult := api.GetLastDownloadFileInfo()
	// Проверяет, есть ли ссылка на файл импорта
	if len(fileResult.FiasCompleteXmlUrl) > 0 {
		// Очищает директорию от ранее скачанных файлов
		is.clearDirectory(false)
		// Получает список названий файлов импорта
		parts := is.getParts()
		// Загружает файл и распаковывает
		xmlFiles := is.directoryService.DownloadAndExtractFile(fileResult.FiasCompleteXmlUrl, "fias_xml.zip", parts...)
		// Читает xml-файлы и импортирует элементы
		cntAddr, cntHouses := is.ParseFiles(xmlFiles)
		// Обновляет версию ФИАС в БД
		versionService.UpdateVersion(is.convertDownloadInfoToVersion(fileResult, cntAddr, cntHouses))
	}

	is.logger.Info("Import finished")
}

// Конвертирует объект файла в объект версии
func (is *ImportService) convertDownloadInfoToVersion(info entity.DownloadFileInfo, cntAddr int, cntHouses int) *versionEntity.Version {
	versionDateSlice := info.TextVersion[len(info.TextVersion)-10 : len(info.TextVersion)]
	versionTime, _ := time.Parse("02.01.2006", versionDateSlice)
	versionDate := versionTime.Format("2006-01-02") + "T00:00:00Z"

	return &versionEntity.Version{
		ID:               info.VersionId,
		FiasVersion:      info.TextVersion,
		UpdateDate:       versionDate,
		RecUpdateAddress: cntAddr,
		RecUpdateHouses:  cntHouses,
	}
}

// Очищает директорию
func (is *ImportService) clearDirectory(force bool) {
	if !is.SkipClear || force {
		is.directoryService.ClearDirectory()
	}
}

// Парсинг файлов и импорт элементов
func (is *ImportService) ParseFiles(files *[]directoryEntity.File) (int, int) {
	var wg sync.WaitGroup
	// Канал подсчета количества адресов
	cha := make(chan int)
	// Канал подсчета количества домов
	chb := make(chan int)
	hasAddress := false
	hasHouse := false
	cntAddr := 0
	cntHouse := 0

	for _, file := range *files {
		// Проверяет наличие файла с адресами
		if r, err := regexp.MatchString(addressEntity.AddressObject{}.GetXmlFile(), file.Path); err == nil && r {
			hasAddress = true
			wg.Add(1)
			// Выполняет импорт адресов
			go is.addressImportService.Import(file.Path, &wg, cha)
		}
		// Проверяет наличие файла с домами
		if r, err := regexp.MatchString(addressEntity.HouseObject{}.GetXmlFile(), file.Path); err == nil && r {
			hasHouse = true
			wg.Add(1)
			// Выполняет импорт домов
			go is.houseImportService.Import(file.Path, &wg, chb)
		}
	}
	if hasAddress {
		cntAddr = <-cha
	}
	if hasHouse {
		cntHouse = <-chb
	}
	wg.Wait()

	return cntAddr, cntHouse
}

// Индексация таблиц БД
func (is *ImportService) Index() {
	var wg sync.WaitGroup
	var guids []string
	// Канал индексации домов при изменении адресов
	indexChan := make(chan addressEntity.IndexObject, is.config.GetInt("workers.houses"))

	// Получает GUID адресов последних загруженных домов
	if !is.IsFull {
		guids = is.houseImportService.GetLastUpdatedGuids(is.Begin)
	}

	wg.Add(2)
	// Индексация таблицы адресов
	go is.addressImportService.Index(is.IsFull, is.Begin, guids, &wg, indexChan)
	// Индексация таблицы домов
	go is.houseImportService.Index(&wg, indexChan)
	wg.Wait()
}
