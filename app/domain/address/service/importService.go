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

type ImportService struct {
	addressImportService *AddressImportService
	houseImportService   *HouseImportService
	logger               interfaces.LoggerInterface
	directoryService     *service.DirectoryService
	config               interfaces.ConfigInterface
	IsFull               bool `default:"false"`
	SkipHouses           bool `default:"false"`
	SkipClear            bool `default:"false"`
	Begin                time.Time
}

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

func (is *ImportService) getParts() []string {
	parts := []string{addressEntity.AddressObject{}.GetXmlFile()}
	if !is.SkipHouses {
		parts = append(parts, addressEntity.HouseObject{}.GetXmlFile())
	}

	return parts
}

func (is *ImportService) CheckUpdates(api *fiasApiService.FiasApiService, versionService *versionService.VersionService, version *versionEntity.Version) {
	result := api.GetAllDownloadFileInfo()
	var needVersionList []entity.DownloadFileInfo
	for _, file := range result {
		if file.VersionId == version.ID {
			break
		}
		needVersionList = append(needVersionList, file)
	}
	parts := is.getParts()

	is.clearDirectory(false)
	if len(needVersionList) == 0 {
		is.logger.Info("Last version is uploaded")
		os.Exit(1)
	}
	for i := len(needVersionList) - 1; i >= 0; i-- {
		uploadedVersion := needVersionList[i]
		cntAddr := 0
		cntHouses := 0

		is.logger.WithFields(interfaces.LoggerFields{
			"version": uploadedVersion,
		}).Debug("Uploaded version info")

		if uploadedVersion.FiasDeltaXmlUrl != "" {
			xmlFiles := is.directoryService.DownloadAndExtractFile(uploadedVersion.FiasDeltaXmlUrl, "fias_delta_xml.zip", parts...)
			cntAddr, cntHouses = is.ParseFiles(xmlFiles)
		}
		is.clearDirectory(true)
		versionService.UpdateVersion(is.convertDownloadInfoToVersion(uploadedVersion, cntAddr, cntHouses))
	}

	is.logger.Info("Import finished")
}

func (is *ImportService) StartFullImport(api *fiasApiService.FiasApiService, versionService *versionService.VersionService) {
	is.IsFull = true
	is.addressImportService.IsFull = true
	is.houseImportService.IsFull = true

	fileResult := api.GetLastDownloadFileInfo()
	if len(fileResult.FiasCompleteXmlUrl) > 0 {
		is.clearDirectory(false)
		parts := is.getParts()
		xmlFiles := is.directoryService.DownloadAndExtractFile(fileResult.FiasCompleteXmlUrl, "fias_xml.zip", parts...)
		cntAddr, cntHouses := is.ParseFiles(xmlFiles)
		versionService.UpdateVersion(is.convertDownloadInfoToVersion(fileResult, cntAddr, cntHouses))
	}

	is.logger.Info("Import finished")
}

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

func (is *ImportService) clearDirectory(force bool) {
	if !is.SkipClear || force {
		err := is.directoryService.ClearDirectory()
		if err != nil {
			is.logger.Fatal(err.Error())
			os.Exit(1)
		}
	}
}

func (is *ImportService) ParseFiles(files *[]directoryEntity.File) (int, int) {
	var wg sync.WaitGroup
	cha := make(chan int)
	chb := make(chan int)
	hasAddress := false
	hasHouse := false
	cntAddr := 0
	cntHouse := 0

	for _, file := range *files {
		if r, err := regexp.MatchString(addressEntity.AddressObject{}.GetXmlFile(), file.Path); err == nil && r {
			hasAddress = true
			wg.Add(1)
			go is.addressImportService.Import(file.Path, &wg, cha)
		}
		if r, err := regexp.MatchString(addressEntity.HouseObject{}.GetXmlFile(), file.Path); err == nil && r {
			hasHouse = true
			wg.Add(1)
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

func (is *ImportService) Index() {
	is.addressImportService.Index(is.IsFull, is.Begin, is.houseImportService.CountAllData(), is.houseImportService.GetByAddressGuid)
}
