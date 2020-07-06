package main

import (
	"context"
	"github.com/olivere/elastic/v7"
	"io"
	"time"
)

var (
	elasticClient *elastic.Client
)

func DoESConnection() {
	var err error

	for {
		elasticClient, err = elastic.NewClient(
			elastic.SetURL("http://"+*host),
			elastic.SetSniff(false),
		)
		if err != nil {
			logPrintln(err)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
}

func GetPrefixIndexName(name string) string {
	return *prefix + name
}

func IndexExists(name string) bool {
	ctx := context.Background()
	exists, err := elasticClient.IndexExists(GetPrefixIndexName(name)).Do(ctx)
	if err != nil {
		logFatal(err)
	}

	return exists
}

func DropIndex(name string) {
	ctx := context.Background()
	if IndexExists(name) {
		indexName := GetPrefixIndexName(name)
		logPrintf("Drop index: %s", indexName)
		_, err := elasticClient.DeleteIndex(indexName).Do(ctx)
		if err != nil {
			logFatal(err)
		}
	}
}

func CreateIndex(name, body string) {
	ctx := context.Background()
	if !IndexExists(name) {
		indexName := GetPrefixIndexName(name)
		logPrintf("Create new index: %s", indexName)
		_, err := elasticClient.CreateIndex(indexName).Body(body).Do(ctx)
		if err != nil {
			logFatal(err)
		}
	}
}

func CreatePreprocessor(pipelineId, pipeline string) {
	logPrintf("Create new preprocessor: %s", pipelineId)
	ctx := context.Background()
	_, err := elasticClient.IngestPutPipeline(pipelineId).BodyString(pipeline).Do(ctx)
	if err != nil {
		logFatal(err)
	}
}

func RefreshIndexes() {
	indexes := []string{GetPrefixIndexName(addressIndexName), GetPrefixIndexName(houseIndexName), GetPrefixIndexName(infoIndexName)}
	elasticClient.Refresh(indexes...)
	elasticClient.Flush(indexes...)
	elasticClient.Forcemerge(indexes...)
	elasticClient.ClearCache(indexes...)
}

func ScrollData(scrollService *elastic.ScrollService) []elastic.SearchHit {
	// Setup a group of goroutines from the excellent errgroup package
	ctx := context.Background()
	scrollService.Scroll("1h")
	var totals []elastic.SearchHit
	maxCount := 0

	for {
		if maxCount > 1000 {
			break
		}
		res, err := scrollService.Do(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			logFatal(err)
		}
		if res == nil || len(res.Hits.Hits) == 0 {
			break
		}
		for _, hit := range res.Hits.Hits {
			totals = append(totals, *hit)
		}
		maxCount++
	}
	scrollService.Clear(ctx)

	return totals
}

func countAllData(index string) int64 {
	res, err := elasticClient.Count(GetPrefixIndexName(index)).Do(context.Background())
	if err != nil {
		logFatal(err)
	}

	return res
}
