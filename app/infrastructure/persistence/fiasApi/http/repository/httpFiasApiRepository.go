package repository

import (
	"encoding/json"
	"github.com/GarinAG/gofias/domain/fiasApi/entity"
	"github.com/GarinAG/gofias/domain/fiasApi/repository"
	"github.com/GarinAG/gofias/infrastructure/persistence/fiasApi/http/dto"
	"github.com/GarinAG/gofias/interfaces"
	"net/http"
)

const (
	httpFiasApiAllFiles = "GetAllDownloadFileInfo"
	httpFiasApiLastFile = "GetLastDownloadFileInfo"
)

type HttpFiasApiRepository struct {
	config interfaces.ConfigInterface
}

func NewHttpFiasApiRepository(config interfaces.ConfigInterface) repository.FiasApiRepositoryInterface {
	return &HttpFiasApiRepository{
		config: config,
	}
}

func (f *HttpFiasApiRepository) GetAllDownloadFileInfo() ([]entity.DownloadFileInfo, error) {
	var files []entity.DownloadFileInfo
	var jsonFiles []dto.JsonDownloadFileInfo

	url := f.config.GetString("fiasApi.url") + httpFiasApiAllFiles
	res, err := f.getHttpClient().Get(url)
	if err != nil {
		return files, err
	}
	json.NewDecoder(res.Body).Decode(&jsonFiles)
	for _, item := range jsonFiles {
		files = append(files, entity.DownloadFileInfo{
			VersionId:          item.VersionId,
			TextVersion:        item.TextVersion,
			FiasCompleteXmlUrl: item.FiasCompleteXmlUrl,
			FiasDeltaXmlUrl:    item.FiasDeltaXmlUrl,
		})
	}

	return files, nil
}

func (f *HttpFiasApiRepository) GetLastDownloadFileInfo() (entity.DownloadFileInfo, error) {
	var file entity.DownloadFileInfo
	var jsonFile dto.JsonDownloadFileInfo
	url := f.config.GetString("fiasApi.url") + httpFiasApiLastFile
	res, err := f.getHttpClient().Get(url)
	if err != nil {
		return file, err
	}
	json.NewDecoder(res.Body).Decode(&jsonFile)
	file = entity.DownloadFileInfo{
		VersionId:          jsonFile.VersionId,
		TextVersion:        jsonFile.TextVersion,
		FiasCompleteXmlUrl: jsonFile.FiasCompleteXmlUrl,
		FiasDeltaXmlUrl:    jsonFile.FiasDeltaXmlUrl,
	}

	return file, nil
}

func (f *HttpFiasApiRepository) getHttpClient() *http.Client {
	return &http.Client{Transport: &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 20,
	}}
}

func (f *HttpFiasApiRepository) convertToEntity(item dto.JsonDownloadFileInfo) entity.DownloadFileInfo {
	return entity.DownloadFileInfo{
		VersionId:          item.VersionId,
		TextVersion:        item.TextVersion,
		FiasCompleteXmlUrl: item.FiasCompleteXmlUrl,
		FiasDeltaXmlUrl:    item.FiasDeltaXmlUrl,
	}
}

func (f *HttpFiasApiRepository) convertToDto(item entity.DownloadFileInfo) dto.JsonDownloadFileInfo {
	return dto.JsonDownloadFileInfo{
		VersionId:          item.VersionId,
		TextVersion:        item.TextVersion,
		FiasCompleteXmlUrl: item.FiasCompleteXmlUrl,
		FiasDeltaXmlUrl:    item.FiasDeltaXmlUrl,
	}
}
