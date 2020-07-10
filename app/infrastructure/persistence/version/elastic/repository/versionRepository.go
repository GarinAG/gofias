package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GarinAG/gofias/domain/version/entity"
	"github.com/GarinAG/gofias/domain/version/repository"
	"github.com/GarinAG/gofias/infrastructure/persistence/version/elastic/dto"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/olivere/elastic/v7"
)

type ElasticVersionRepository struct {
	elasticClient *elastic.Client
	indexName     string
}

func NewElasticVersionRepository(elasticClient *elastic.Client, configInterface interfaces.ConfigInterface) repository.VersionRepositoryInterface {
	return &ElasticVersionRepository{
		elasticClient: elasticClient,
		indexName:     configInterface.GetString("project.prefix") + entity.Version{}.TableName(),
	}
}

func (v *ElasticVersionRepository) GetVersion() (*entity.Version, error) {
	versionSearchResult, _ := v.elasticClient.Search(v.indexName).
		Sort("version_id", false).
		Size(1).
		RequestCache(false).
		Do(context.Background())

	var dtoItem dto.JsonVersionDto
	if versionSearchResult != nil && len(versionSearchResult.Hits.Hits) > 0 {
		if err := json.Unmarshal(versionSearchResult.Hits.Hits[0].Source, &dtoItem); err != nil {
			return nil, err
		}

		return v.convertToEntity(dtoItem), nil
	}

	return nil, nil
}

func (v *ElasticVersionRepository) SetVersion(version *entity.Version) error {
	res, err := v.elasticClient.Bulk().
		Index(v.indexName).
		Refresh("true").
		Add(elastic.NewBulkIndexRequest().Doc(v.convertToDto(*version))).
		Do(context.Background())

	if err != nil {
		return err
	}
	if res.Errors {
		return errors.New("Bulk commit failed")
	}

	return nil
}

func (v *ElasticVersionRepository) convertToEntity(item dto.JsonVersionDto) *entity.Version {
	return &entity.Version{
		ID:               item.ID,
		FiasVersion:      item.FiasVersion,
		UpdateDate:       item.UpdateDate,
		RecUpdateAddress: item.RecUpdateAddress,
		RecUpdateHouses:  item.RecUpdateHouses,
	}
}

func (v *ElasticVersionRepository) convertToDto(item entity.Version) *dto.JsonVersionDto {
	return &dto.JsonVersionDto{
		ID:               item.ID,
		FiasVersion:      item.FiasVersion,
		UpdateDate:       item.UpdateDate,
		RecUpdateAddress: item.RecUpdateAddress,
		RecUpdateHouses:  item.RecUpdateHouses,
	}
}
