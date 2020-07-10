package service

import (
	addressEntity "github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	addressService "github.com/GarinAG/gofias/domain/address/service"
	directoryEntity "github.com/GarinAG/gofias/domain/directory/entity"
	"github.com/GarinAG/gofias/domain/directory/service"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/spf13/viper"
	"regexp"
	"sync"
)

type ImportService struct {
	addressImportService *addressService.AddressImportService
	houseImportService   *addressService.HouseImportService
	logger               interfaces.LoggerInterface
	directoryService     *service.DirectoryService
	isFull               bool `default:"false"`
}

func NewImportService(logger interfaces.LoggerInterface, ds *service.DirectoryService, addressImportService *addressService.AddressImportService, houseImportService *addressService.HouseImportService) *ImportService {
	return &ImportService{
		addressImportService: addressImportService,
		houseImportService:   houseImportService,
		logger:               logger,
		directoryService:     ds,
	}
}

func (is *ImportService) CheckUpdates(api *fiasApiService.FiasApiService, version int) {
	result := api.GetAllDownloadFileInfo()
	for _, file := range result {
		if file.VersionId > version {
			xmlFiles := is.directoryService.DownloadAndExtractFile(file.FiasDeltaXmlUrl)
			is.ParseFiles(xmlFiles)
		}
	}
}

func (is *ImportService) StartFullImport(api *fiasApiService.FiasApiService) {
	is.isFull = true
	fileResult := api.GetLastDownloadFileInfo()
	if len(fileResult.FiasCompleteXmlUrl) > 0 {
		xmlFiles := is.directoryService.DownloadAndExtractFile(fileResult.FiasCompleteXmlUrl)
		is.ParseFiles(xmlFiles)
	}
}

func (is *ImportService) ParseFiles(files *[]directoryEntity.File) {
	var wg sync.WaitGroup
	for _, file := range *files {
		if r, err := regexp.MatchString(addressEntity.AddressObject{}.GetXmlFile(), file.Path); err == nil && r {
			wg.Add(1)
			go is.addressImportService.Import(file.Path, &wg, is.isFull)
		}
		if r, err := regexp.MatchString(addressEntity.HouseObject{}.GetXmlFile(), file.Path); err == nil && r {
			wg.Add(1)
			go is.houseImportService.Import(file.Path, &wg, is.isFull)
		}
	}
	wg.Wait()

}

func insertCollection(repo repository.InsertUpdateInterface, collection []interface{}, node interface{}, isFull bool) []interface{} {
	if collection == nil {
		collection = append(collection, node)
		return collection
	}
	if node == nil {
		repo.InsertUpdateCollection(collection, isFull)
		return collection[:0]
	}
	if len(collection) < viper.GetInt("import.collectionCount") {
		collection = append(collection, node)
		return collection
	} else {
		collection = append(collection, node)
		repo.InsertUpdateCollection(collection, isFull)
		return collection[:0]
	}
}
