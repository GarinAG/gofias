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

// Репозиторий версий http
type HttpFiasApiRepository struct {
	config interfaces.ConfigInterface // Конфигурация
}

// Инициализация репозитория
func NewHttpFiasApiRepository(config interfaces.ConfigInterface) repository.FiasApiRepositoryInterface {
	return &HttpFiasApiRepository{
		config: config,
	}
}

// Получить все версии БД ФИАС
func (f *HttpFiasApiRepository) GetAllDownloadFileInfo() ([]entity.DownloadFileInfo, error) {
	var files []entity.DownloadFileInfo
	var jsonFiles []dto.JsonDownloadFileInfo

	url := f.config.GetString("fiasApi.url") + httpFiasApiAllFiles
	res, err := f.getHttpClient().Get(url)
	if err != nil {
		return files, err
	}
	// Конвертирует структуру ответа в DTO
	err = json.NewDecoder(res.Body).Decode(&jsonFiles)
	if err != nil {
		return files, err
	}
	for _, item := range jsonFiles {
		// Конвертирует DTO в объект версии
		files = append(files, f.convertToEntity(item))
	}

	return files, nil
}

// Получить последнюю версию БД ФИАС
func (f *HttpFiasApiRepository) GetLastDownloadFileInfo() (entity.DownloadFileInfo, error) {
	var file entity.DownloadFileInfo
	var jsonFile dto.JsonDownloadFileInfo
	url := f.config.GetString("fiasApi.url") + httpFiasApiLastFile
	res, err := f.getHttpClient().Get(url)
	if err != nil {
		return file, err
	}
	// Конвертирует структуру ответа в DTO
	err = json.NewDecoder(res.Body).Decode(&jsonFile)
	if err != nil {
		return file, err
	}
	// Конвертирует DTO в объект версии
	file = f.convertToEntity(jsonFile)

	return file, nil
}

// Инициализация http-клиента
func (f *HttpFiasApiRepository) getHttpClient() *http.Client {
	return &http.Client{Transport: &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 20,
	}}
}

// Конвертирует DTO в объект версии
func (f *HttpFiasApiRepository) convertToEntity(item dto.JsonDownloadFileInfo) entity.DownloadFileInfo {
	return entity.DownloadFileInfo{
		VersionId:          item.VersionId,
		TextVersion:        item.TextVersion,
		FiasCompleteXmlUrl: item.FiasCompleteXmlUrl,
		FiasDeltaXmlUrl:    item.FiasDeltaXmlUrl,
	}
}

// Конвертирует объект версии в DTO
func (f *HttpFiasApiRepository) convertToDto(item entity.DownloadFileInfo) dto.JsonDownloadFileInfo {
	return dto.JsonDownloadFileInfo{
		VersionId:          item.VersionId,
		TextVersion:        item.TextVersion,
		FiasCompleteXmlUrl: item.FiasCompleteXmlUrl,
		FiasDeltaXmlUrl:    item.FiasDeltaXmlUrl,
	}
}
