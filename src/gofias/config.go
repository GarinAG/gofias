package main

import "flag"

const (
	addressIndexName = "address"
	houseIndexName   = "houses"
	infoIndexName    = "info"
	repositoryName   = "fias"
	snapshotName     = "full"

	addrFilePart = "AS_ADDROBJ_*"
	addrPipeline = "addr_drop_pipeline"
	addrTag      = "Object"

	housesFilePart = "AS_HOUSE_*"
	housesPipeline = "house_drop_pipeline"
	housesTag      = "House"

	fiasXml        = "fias_xml.zip"
	fiasDeltaXml   = "fias_delta_xml.zip"
	fiasUrl        = "https://fias.nalog.ru/Public/Downloads/Actual/"
	fiasServiceUrl = "https://fias.nalog.ru/WebServices/Public/DownloadService.asmx"

	urlFullPath = fiasUrl + fiasXml

	dateTimeZone = "T00:00:00Z"
)

var (
	storage    = flag.String("storage", "/usr/share/elasticsearch/snapshots", "Snapshots storage path")
	host       = flag.String("host", "localhost:9200", "Elasticsearch host")
	prefix     = flag.String("prefix", "fias_", "Prefix for elasticsearch indexes")
	tmp        = flag.String("tmp", "/tmp/fias/", "Tmp folder relative path in user home dir")
	status     = flag.Bool("status", false, "Show import status")
	force      = flag.Bool("force", false, "Force full download")
	bulkSize   = flag.Int("bulk-size", 1000, "Number of documents to collect before committing")
	skipHouses = flag.Bool("skip-houses", false, "Skip houses index")
)
