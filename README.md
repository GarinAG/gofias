# Golang fias to elasticsearch import

gofias is a [Go](http://www.golang.org/) (golang) library that import [fias](https://fias.nalog.ru/) data to [Elasticsearch](http://www.elasticsearch.org/)

## Usage

```
git clone https://github.com/GarinAG/gofias.git
cd gofias
go build ./src/gofias/
./gofias --host=localhost:9200 --status=1 --skip-houses
```

## CLI props

* `bulk-size (int)` - Number of documents to collect before committing (default `1000`)
* `force (bool)` - Force full download (default `false`)
* `host (string)` - Elasticsearch host (default `"localhost:9200"`)
* `prefix (string)` - Prefix for elasticsearch indexes (default `"fias_"`)
* `skip-houses (bool)` - Skip houses index (default `false`)
* `status (bool)` - Show import status (default `false`)
* `storage (string)` - Snapshots storage path (default `"/usr/share/elasticsearch/snapshots"`)
* `tmp (string)` - Tmp folder relative path in user home dir (default `"/tmp/fias/"`)
