package main

import (
	"context"
	elastic "github.com/olivere/elastic/v7"
	"io"
	"log"
	"time"
)

var (
	elasticClient *elastic.Client
)

func doESConnection() {
	var err error

	for {
		elasticClient, err = elastic.NewClient(
			elastic.SetURL("http://"+*host),
			elastic.SetSniff(false),
		)
		if err != nil {
			log.Println(err)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
}

func getPrefixIndexName(name string) string {
	return *prefix + name
}

func indexExists(name string) bool {
	ctx := context.Background()
	exists, err := elasticClient.IndexExists(getPrefixIndexName(name)).Do(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return exists
}

func dropIndex(name string) {
	ctx := context.Background()
	if indexExists(name) {
		indexName := getPrefixIndexName(name)
		log.Printf("Drop index: %s", indexName)
		_, err := elasticClient.DeleteIndex(indexName).Do(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func createIndex(name, body string) {
	ctx := context.Background()
	if !indexExists(name) {
		indexName := getPrefixIndexName(name)
		log.Printf("Create new index: %s", indexName)
		_, err := elasticClient.CreateIndex(indexName).Body(body).Do(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func createPreprocessor(pipelineId, pipeline string) {
	log.Printf("Create new preprocessor: %s", pipelineId)
	ctx := context.Background()
	_, err := elasticClient.IngestPutPipeline(pipelineId).BodyString(pipeline).Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func refreshIndexes() {
	elasticClient.Refresh()
	elasticClient.Flush()
	elasticClient.Forcemerge()
}

func scrollData(scrollService *elastic.ScrollService) []elastic.SearchHit {
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
			log.Fatal(err)
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
	res, err := elasticClient.Count(getPrefixIndexName(index)).Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	return res
}
