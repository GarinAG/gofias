package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GarinAG/gofias/domain/version/entity"
	"github.com/GarinAG/gofias/domain/version/repository"
	elasticHelper "github.com/GarinAG/gofias/infrastructure/persistence/elastic"
	"github.com/GarinAG/gofias/infrastructure/persistence/version/elastic/dto"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/olivere/elastic/v7"
	"strconv"
)

const (
	// Структура индекса в эластике
	versionIndexSettings = `
	{
	  "settings": {
		"index": {
		  "number_of_shards": 1,
		  "number_of_replicas": "0",
		  "refresh_interval": "-1",
		  "requests": {
			"cache": {
			  "enable": "false"
			}
		  },
		  "blocks": {
			"read_only_allow_delete": "false"
		  }
		}
	  },
	  "mappings": {
		"dynamic": false,
		"properties": {
		  "version_id": {
			"type": "integer"
		  },
		  "fias_version": {
			"type": "keyword"
		  },
		  "update_date": {
			"type": "date"
		  },
		  "rec_upd_address": {
			"type": "integer"
		  },
		  "rec_upd_houses": {
			"type": "integer"
		  }
		}
	  }
	}
	`
)

// Репозиторий версий в эластике
type ElasticVersionRepository struct {
	elasticClient *elasticHelper.Client // Клиент эластика
	indexName     string                // Название индекса
}

// Инициализация репозитория
func NewElasticVersionRepository(elasticClient *elasticHelper.Client, configInterface interfaces.ConfigInterface) repository.VersionRepositoryInterface {
	repos := &ElasticVersionRepository{
		elasticClient: elasticClient,
		indexName:     configInterface.GetString("project.prefix") + entity.Version{}.TableName(),
	}

	return repos
}

// Инициализация индекса
func (v *ElasticVersionRepository) Init() error {
	return v.elasticClient.CreateIndex(v.indexName, versionIndexSettings)
}

// Получить текущую версию БД ФИАС
func (v *ElasticVersionRepository) GetVersion() (*entity.Version, error) {
	versionSearchResult, _ := v.elasticClient.Client.Search(v.indexName).
		Sort("version_id", false).
		Size(1).
		RequestCache(false).
		Do(context.Background())

	var dtoItem dto.JsonVersionDto
	// Конвертирует структуру ответа в DTO
	if versionSearchResult != nil && len(versionSearchResult.Hits.Hits) > 0 {
		if err := json.Unmarshal(versionSearchResult.Hits.Hits[0].Source, &dtoItem); err != nil {
			return nil, err
		}

		return v.convertToEntity(dtoItem), nil
	}

	return nil, nil
}

// Сохранить версию
func (v *ElasticVersionRepository) SetVersion(version *entity.Version) error {
	doc := v.convertToDto(*version)
	id := strconv.Itoa(doc.ID)
	res, err := v.elasticClient.Client.Bulk().
		Index(v.indexName).
		Refresh("true").
		Add(elastic.NewBulkIndexRequest().Id(id).Doc(doc)).
		Do(context.Background())

	if err != nil {
		return err
	}
	if res.Errors {
		return errors.New("Bulk commit failed")
	}

	return nil
}

// Конвертирует объект версии эластика в объект версии
func (v *ElasticVersionRepository) convertToEntity(item dto.JsonVersionDto) *entity.Version {
	return &entity.Version{
		ID:               item.ID,
		FiasVersion:      item.FiasVersion,
		UpdateDate:       item.UpdateDate,
		RecUpdateAddress: item.RecUpdateAddress,
		RecUpdateHouses:  item.RecUpdateHouses,
	}
}

// Конвертирует объект версии в объект версии эластика
func (v *ElasticVersionRepository) convertToDto(item entity.Version) *dto.JsonVersionDto {
	return &dto.JsonVersionDto{
		ID:               item.ID,
		FiasVersion:      item.FiasVersion,
		UpdateDate:       item.UpdateDate,
		RecUpdateAddress: item.RecUpdateAddress,
		RecUpdateHouses:  item.RecUpdateHouses,
	}
}

// Удалить индекс
func (v *ElasticVersionRepository) Clear() error {
	return v.elasticClient.DropIndex(v.indexName)
}
