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
	bulkSize     = flag.Int("bulk-size", 1000, "Number of documents to collect before committing")
	numCpu       = flag.Int("cpu", 0, "Count of CPU usage")
	force        = flag.Bool("force", false, "Force full download")
	forceIndex   = flag.Bool("force-index", false, "Start force index without import")
	host         = flag.String("host", "localhost:9200", "Elasticsearch host")
	logPath      = flag.String("logs", "./logs", "Logs dir path")
	prefix       = flag.String("prefix", "fias_", "Prefix for elasticsearch indexes")
	skipClear    = flag.Bool("skip-clear", false, "Skip clear tmp directory on start")
	skipHouses   = flag.Bool("skip-houses", false, "Skip houses index")
	skipSnapshot = flag.Bool("skip-snapshot", false, "Skip create ElasticSearch snapshot")
	status       = flag.Bool("status", false, "Show import status")
	storage      = flag.String("storage", "/usr/share/elasticsearch/snapshots", "Snapshots storage path")
	tmp          = flag.String("tmp", "/tmp/fias/", "Tmp folder relative path in user home dir")
)
