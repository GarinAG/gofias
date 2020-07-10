package elastic

import (
	"context"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/olivere/elastic/v7"
	"io"
)

func InitElasticClient(configInterface interfaces.ConfigInterface) *elastic.Client {
	elasticClient, err := elastic.NewClient(
		elastic.SetURL("http://"+configInterface.GetString("elastic.host")),
		elastic.SetSniff(false),
	)

	if err != nil {
		panic(err)
	}

	return elasticClient
}

func IndexExists(elasticClient *elastic.Client, index string) (bool, error) {
	ctx := context.Background()
	exists, err := elasticClient.IndexExists(index).Do(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func DropIndex(elasticClient *elastic.Client, index string) error {
	ctx := context.Background()
	exists, err := IndexExists(elasticClient, index)
	if err != nil {
		return err
	}
	if exists {
		_, err := elasticClient.DeleteIndex(index).Do(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateIndex(elasticClient *elastic.Client, index string, body string) error {
	ctx := context.Background()
	exists, err := IndexExists(elasticClient, index)
	if err != nil {
		return err
	}
	if !exists {
		_, err := elasticClient.CreateIndex(index).Body(body).Do(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreatePreprocessor(elasticClient *elastic.Client, pipelineId, pipeline string) error {
	ctx := context.Background()
	_, err := elasticClient.IngestPutPipeline(pipelineId).BodyString(pipeline).Do(ctx)
	if err != nil {
		return err
	}

	return nil
}

func RefreshIndexes(elasticClient *elastic.Client, indexes []string) {
	elasticClient.Refresh(indexes...)
	elasticClient.Flush(indexes...)
	elasticClient.Forcemerge(indexes...)
	elasticClient.ClearCache(indexes...)
}

func ScrollData(scrollService *elastic.ScrollService) ([]elastic.SearchHit, error) {
	ctx := context.Background()
	scrollService.Scroll("1h")
	var totals []elastic.SearchHit

	for {
		res, err := scrollService.Do(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if res == nil || len(res.Hits.Hits) == 0 {
			break
		}
		for _, hit := range res.Hits.Hits {
			totals = append(totals, *hit)
		}
	}

	return totals, nil
}

func countAllData(elasticClient *elastic.Client, index string) (int64, error) {
	res, err := elasticClient.Count(index).Do(context.Background())
	if err != nil {
		return 0, err
	}

	return res, nil
}
