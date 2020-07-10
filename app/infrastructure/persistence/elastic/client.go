package elastic

import (
	"context"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/olivere/elastic/v7"
	"io"
)

type Client struct {
	Client *elastic.Client
}

func NewElasticClient(configInterface interfaces.ConfigInterface) *Client {
	scheme := configInterface.GetString("elastic.scheme")
	user := configInterface.GetString("elastic.username")
	pass := configInterface.GetString("elastic.password")

	if scheme == "" {
		scheme = "http"
	}
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(scheme + "://" + configInterface.GetString("elastic.host")),
		elastic.SetSniff(configInterface.GetBool("elastic.sniff")),
		elastic.SetGzip(configInterface.GetBool("elastic.gzip")),
	}
	if user != "" && pass != "" {
		options = append(options, elastic.SetBasicAuth(user, pass))
	}

	client, err := elastic.NewClient(options...)

	if err != nil {
		panic(err)
	}

	return &Client{
		Client: client,
	}
}

func (e *Client) IndexExists(index string) (bool, error) {
	ctx := context.Background()
	exists, err := e.Client.IndexExists(index).Do(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (e *Client) DropIndex(index string) error {
	ctx := context.Background()
	exists, err := e.IndexExists(index)
	if err != nil {
		return err
	}
	if exists {
		_, err := e.Client.DeleteIndex(index).Do(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Client) CreateIndex(index string, body string) error {
	ctx := context.Background()
	exists, err := e.IndexExists(index)
	if err != nil {
		return err
	}
	if !exists {
		_, err := e.Client.CreateIndex(index).Body(body).Do(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Client) CreatePreprocessor(pipelineId, pipeline string) error {
	ctx := context.Background()
	_, err := e.Client.IngestPutPipeline(pipelineId).BodyString(pipeline).Do(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (e *Client) RefreshIndexes(indexes []string) {
	e.Client.Refresh(indexes...)
	e.Client.Flush(indexes...)
	e.Client.Forcemerge(indexes...)
	e.Client.ClearCache(indexes...)
}

func (e *Client) ScrollData(scrollService *elastic.ScrollService) ([]elastic.SearchHit, error) {
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

func (e *Client) CountAllData(index string) (int64, error) {
	res, err := e.Client.Count(index).Do(context.Background())
	if err != nil {
		return 0, err
	}

	return res, nil
}
