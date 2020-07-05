package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/olivere/elastic/v7"
	"github.com/tiaguinho/gosoap"
	"log"
	"strconv"
	"time"
)

const (
	infoIndexSettings = `
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
			"type": "keyword"
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

type VersionItemElastic struct {
	ID               string `json:"version_id"`
	FiasVersion      string `json:"fias_version"`
	UpdateDate       string `json:"update_date"`
	RecUpdateAddress string `json:"rec_upd_address"`
	RecUpdateHouses  string `json:"rec_upd_houses"`
}

type DownloadFileInfo struct {
	VersionId          string `xml:"VersionId"`
	TextVersion        string `xml:"TextVersion"`
	FiasCompleteXmlUrl string `xml:"FiasCompleteXmlUrl"`
	FiasDeltaXmlUrl    string `xml:"FiasDeltaXmlUrl"`
}

type GetLastDownloadFileInfoResponse struct {
	Result DownloadFileInfo `xml:"GetLastDownloadFileInfoResult"`
}

type GetAllDownloadFileInfoResponse struct {
	Result GetAllDownloadFileInfoResult `xml:"GetAllDownloadFileInfoResult"`
}

type GetAllDownloadFileInfoResult struct {
	Result []DownloadFileInfo `xml:"DownloadFileInfo"`
}

var (
	currentVersion               VersionItemElastic
	lastDownloadVersion          DownloadFileInfo
	downloadVersionList          []DownloadFileInfo
	lastDownloadFileInfoResponse GetLastDownloadFileInfoResponse
	allDownloadFileInfoResponse  GetAllDownloadFileInfoResponse
	recUpdateAddress             uint64 = 0
	recUpdateHouses              uint64 = 0
	isUpdate                     bool
	versionDate                  string
)

func getLastVersion() {
	indexName := getPrefixIndexName(infoIndexName)
	createIndex(infoIndexName, infoIndexSettings)

	versionSearchResult, _ := elasticClient.Search(indexName).
		Sort("version_id", false).
		Size(1).
		Do(context.Background())

	var item VersionItemElastic
	if versionSearchResult != nil && len(versionSearchResult.Hits.Hits) > 0 {
		if err := json.Unmarshal(versionSearchResult.Hits.Hits[0].Source, &item); err != nil {
			log.Fatal(err)
		}
	} else {
		item = VersionItemElastic{}
	}

	currentVersion = item
}

func updateInfo(version DownloadFileInfo) {
	currentVersion = VersionItemElastic{
		ID:               version.VersionId,
		FiasVersion:      version.TextVersion,
		RecUpdateAddress: strconv.FormatUint(recUpdateAddress, 10),
		RecUpdateHouses:  strconv.FormatUint(recUpdateHouses, 10),
		UpdateDate:       time.Now().Format("2006-01-02T00:00:00"),
	}
	log.Printf("Save version info: %s %s", version.VersionId, version.TextVersion)

	res, err := elasticClient.Bulk().
		Index(getPrefixIndexName(infoIndexName)).
		Refresh("true").
		Add(elastic.NewBulkIndexRequest().Doc(currentVersion)).
		Do(context.Background())

	if err != nil {
		log.Fatal(err)
	}
	if res.Errors {
		log.Fatal("Bulk commit failed")
	}
}

func getLastDownloadVersion() {
	res := executeSoap("GetLastDownloadFileInfo", gosoap.Params{})
	if err := xml.Unmarshal(res.Body, &lastDownloadFileInfoResponse); err != nil {
		log.Fatalf("xml.Unmarshal error: %s", err)
	}
	lastDownloadVersion = lastDownloadFileInfoResponse.Result
}

func getDownloadVersionList() {
	res := executeSoap("GetAllDownloadFileInfo", gosoap.Params{})
	if err := xml.Unmarshal(res.Body, &allDownloadFileInfoResponse); err != nil {
		log.Fatalf("xml.Unmarshal error: %s", err)
	}

	downloadVersionList = allDownloadFileInfoResponse.Result.Result
}
