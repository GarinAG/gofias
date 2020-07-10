package service

import (
	addressEntity "github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	addressService "github.com/GarinAG/gofias/domain/address/service"
	directoryEntity "github.com/GarinAG/gofias/domain/directory/entity"
	"github.com/GarinAG/gofias/domain/directory/service"
	"github.com/GarinAG/gofias/domain/fiasApi/entity"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	"github.com/GarinAG/gofias/interfaces"
	"os"
	"regexp"
	"sync"
	"time"
)

type ImportService struct {
	addressImportService *addressService.AddressImportService
	houseImportService   *addressService.HouseImportService
	logger               interfaces.LoggerInterface
	directoryService     *service.DirectoryService
	config               interfaces.ConfigInterface
	IsFull               bool `default:"false"`
	SkipHouses           bool `default:"false"`
	SkipClear            bool `default:"false"`
}

func NewImportService(logger interfaces.LoggerInterface, ds *service.DirectoryService, addressImportService *addressService.AddressImportService, houseImportService *addressService.HouseImportService, config interfaces.ConfigInterface) *ImportService {
	return &ImportService{
		addressImportService: addressImportService,
		houseImportService:   houseImportService,
		logger:               logger,
		directoryService:     ds,
		config:               config,
		IsFull:               false,
	}
}

func (is *ImportService) CheckUpdates(api *fiasApiService.FiasApiService, version int) {
	result := api.GetAllDownloadFileInfo()
	var needVersionList []entity.DownloadFileInfo
	for _, file := range result {
		if file.VersionId == version {
			break
		}
		needVersionList = append(needVersionList, file)
	}
	parts := []string{addressEntity.HouseObject{}.GetXmlFile()}
	if !is.SkipHouses {
		parts = append(parts, addressEntity.HouseObject{}.GetXmlFile())
	}

	is.clearDirectory(false)
	for i := len(needVersionList) - 1; i >= 0; i-- {
		xmlFiles := is.directoryService.DownloadAndExtractFile(needVersionList[i].FiasDeltaXmlUrl, "fias_delta.zip", parts...)
		is.ParseFiles(xmlFiles)
		is.clearDirectory(true)
	}
}

func (is *ImportService) StartFullImport(api *fiasApiService.FiasApiService) {
	is.IsFull = true
	fileResult := api.GetLastDownloadFileInfo()
	if len(fileResult.FiasCompleteXmlUrl) > 0 {
		is.clearDirectory(false)
		parts := []string{addressEntity.HouseObject{}.GetXmlFile()}
		if !is.SkipHouses {
			parts = append(parts, addressEntity.HouseObject{}.GetXmlFile())
		}
		xmlFiles := is.directoryService.DownloadAndExtractFile(fileResult.FiasCompleteXmlUrl, "fias_full.zip", parts...)
		is.ParseFiles(xmlFiles)
	}
}

func (is *ImportService) ParseFiles(files *[]directoryEntity.File) {
	var wg sync.WaitGroup
	begin := time.Now()

	for _, file := range *files {
		if r, err := regexp.MatchString(addressEntity.AddressObject{}.GetXmlFile(), file.Path); err == nil && r {
			wg.Add(1)
			go is.addressImportService.Import(file.Path, &wg, is.IsFull, is.config.GetInt("bach.size"), is.insertCollection)
		}
		if r, err := regexp.MatchString(addressEntity.HouseObject{}.GetXmlFile(), file.Path); err == nil && r {
			wg.Add(1)
			go is.houseImportService.Import(file.Path, &wg, is.IsFull, is.config.GetInt("bach.size"), is.insertCollection)
		}
	}
	wg.Wait()
	is.Index(begin)
}

func (is *ImportService) Index(begin time.Time) {
	is.addressImportService.Index(is.houseImportService.GetRepo(), is.IsFull, begin)
}

func (is *ImportService) insertCollection(repo repository.InsertUpdateInterface, collection []interface{}, node interface{}, isFull bool, size int) []interface{} {
	if collection == nil {
		collection = append(collection, node)
		return collection
	}
	if node == nil {
		err := repo.InsertUpdateCollection(collection, isFull)
		if err != nil {
			is.logger.Error(err.Error())
		}
		return collection[:0]
	}
	if len(collection) < size {
		collection = append(collection, node)
		return collection
	} else {
		collection = append(collection, node)
		err := repo.InsertUpdateCollection(collection, isFull)
		if err != nil {
			is.logger.Error(err.Error())
		}
		return collection[:0]
	}
}

func (is *ImportService) clearDirectory(force bool) {
	if !is.SkipClear || force {
		err := is.directoryService.ClearDirectory()
		if err != nil {
			is.logger.Fatal(err.Error())
			os.Exit(1)
		}
	}
}
