# GoFias

A [Go](http://www.golang.org/) (golang) library that import [fias](https://fias.nalog.ru/) data to [Elasticsearch](http://www.elasticsearch.org/)

Read this in other languages: [Русский](README.md), [English](README.en.md).

## Base CLI props
* `config-path (string)` - Set additional config path (default `./`)
* `config-type (string)` - Set additional config type (default `yaml`)

## FIAS import service usage
```shell script
cd GOROOT/src/gofias/app/
go build -o ./fias ./application/cli/
cd ..
./fias update --skip-houses --skip-clear
```

## FIAS import CLI props
* `skip-clear (bool)` - Skip clear tmp directory on start (default `false`)
* `skip-houses (bool)` - Skip houses index (default `false`)
* `skip-osm (bool)` - Skip geo-data import (default `false`)

## FIAS grpc server usage

### With docker-compose
```shell script
cd GOROOT/src/gofias
docker-compose up -d
```

### Without docker-compose
```shell script
cd GOROOT/src/gofias
go build -o ./grpc_fias ./application/grpc/
cd ..
./grpc_fias
```

## ElasticSearch indexes info

### address

Contains information about FIAS addresses

<details><summary>Index mapping</summary>
<p>

```json
{
  "settings": {
    "index": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "refresh_interval": "5s",
      "requests": {
        "cache": {
          "enable": "true"
        }
      },
      "blocks": {
        "read_only_allow_delete": "false"
      },
      "analysis": {
        "filter": {
          "russian_stemmer": {
            "type": "stemmer",
            "name": "russian"
          },
          "edge_ngram": {
            "type": "edge_ngram",
            "min_gram": "1",
            "max_gram": "40"
          }
        },
        "analyzer": {
          "edge_ngram_analyzer": {
            "filter": [
              "lowercase",
              "edge_ngram"
            ],
            "tokenizer": "standard"
          },
          "keyword_analyzer": {
            "filter": [
              "lowercase"
            ],
            "tokenizer": "standard"
          }
        }
      }
    }
  },
  "mappings": {
    "dynamic": false,
    "properties": {
      "address_suggest": {
        "type": "text",
        "analyzer": "edge_ngram_analyzer",
        "search_analyzer": "keyword_analyzer"
      },
      "full_address": {
        "type": "keyword"
      },
      "formal_name": {
        "type": "keyword"
      },
      "full_name": {
        "type": "text",
        "analyzer": "edge_ngram_analyzer",
        "search_analyzer": "keyword_analyzer",
        "fields": {
          "keyword": {
            "type": "keyword"
          }
        }
      },
      "ao_id": {
        "type": "keyword"
      },
      "ao_guid": {
        "type": "keyword"
      },
      "parent_guid": {
        "type": "keyword"
      },
      "ao_level": {
        "type": "integer"
      },
      "code": {
        "type": "keyword"
      },
      "short_name": {
        "type": "keyword"
      },
      "off_name": {
        "type": "keyword"
      },
      "curr_status": {
        "type": "integer"
      },
      "act_status": {
        "type": "integer"
      },
      "live_status": {
        "type": "integer"
      },
      "postal_code": {
        "type": "keyword"
      },
      "region_code": {
        "type": "keyword"
      },
      "district_guid": {
        "type": "keyword"
      },
      "district": {
        "type": "keyword"
      },
      "district_type": {
        "type": "keyword"
      },
      "district_full": {
        "type": "keyword"
      },
      "settlement_guid": {
        "type": "keyword"
      },
      "settlement": {
        "type": "keyword"
      },
      "settlement_type": {
        "type": "keyword"
      },
      "settlement_full": {
        "type": "keyword"
      },
      "street": {
        "type": "keyword"
      },
      "street_type": {
        "type": "keyword"
      },
      "street_full": {
        "type": "keyword"
      },
      "okato": {
        "type": "keyword"
      },
      "oktmo": {
        "type": "keyword"
      },
      "start_date": {
        "type": "date"
      },
      "end_date": {
        "type": "date"
      },
      "bazis_update_date": {
        "type": "date"
      },
      "update_date": {
        "type": "date"
      },
      "location": {
        "type": "geo_point",
        "ignore_malformed": true
      },
      "houses": {
        "type": "nested",
        "properties": {
          "house_id": {
            "type": "keyword"
          },
          "house_full_num": {
            "type": "keyword"
          }
        }
      }
    }
  }
}
```

</p>
</details>

<details><summary>Index pipeline</summary>
<p>

```json
{
  "description":
  "drop not actual addresses",
  "processors": [{
    "drop": {
      "if": "ctx.curr_status != 0"
    }
  }, {
    "drop": {
      "if": "ctx.act_status != 1"
    }
  }, {
    "drop": {
      "if": "ctx.live_status != 1"
    }
  }]
}
```

</p>
</details>

### houses

Contains information about FIAS houses

<details><summary>Index mapping</summary>
<p>

```json
{
  "settings": {
    "index": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "refresh_interval": "5s",
      "requests": {
        "cache": {
          "enable": "true"
        }
      },
      "blocks": {
        "read_only_allow_delete": "false"
      },
      "analysis": {
        "filter": {
          "russian_stemmer": {
            "type": "stemmer",
            "name": "russian"
          },
          "ngram": {
            "type": "ngram",
            "min_gram": "1",
            "max_gram": "15"
          },
          "edge_ngram": {
            "type": "edge_ngram",
            "min_gram": "1",
            "max_gram": "50"
          }
        },
        "analyzer": {
          "ngram_analyzer": {
            "filter": ["lowercase", "ngram"],
            "tokenizer": "standard"
          },
          "edge_ngram_analyzer": {
            "filter": ["lowercase", "edge_ngram"],
            "tokenizer": "standard"
          },
          "keyword_analyzer": {
            "filter": ["lowercase"],
            "tokenizer": "standard"
          }
        }
      },
      "max_ngram_diff": 14
    }
  },
  "mappings": {
    "dynamic": false,
    "properties": {
      "house_id": {
        "type": "keyword"
      },
      "house_guid": {
        "type": "keyword"
      },
      "ao_guid": {
        "type": "keyword"
      },
      "build_num": {
        "type": "keyword"
      },
      "house_num": {
        "type": "keyword"
      },
      "address_suggest": {
        "type": "text",
        "analyzer": "edge_ngram_analyzer",
        "search_analyzer": "keyword_analyzer"
      },
      "house_full_num": {
        "type": "text",
        "analyzer": "ngram_analyzer",
        "search_analyzer": "keyword_analyzer",
        "fields": {
          "keyword": {
            "type": "keyword"
          }
        }
      },
      "full_address": {
        "type": "keyword"
      },
      "str_num": {
        "type": "keyword"
      },
      "postal_code": {
        "type": "keyword"
      },
      "counter": {
        "type": "keyword"
      },
      "end_date": {
        "type": "date"
      },
      "start_date": {
        "type": "date"
      },
      "bazis_update_date": {
        "type": "date"
      },
      "update_date": {
        "type": "date"
      },
      "cad_num": {
        "type": "keyword"
      },
      "okato": {
        "type": "keyword"
      },
      "oktmo": {
        "type": "keyword"
      },
      "location": {
        "type": "geo_point",
        "ignore_malformed": true
      }
    }
  }
}  
```

</p>
</details>

<details><summary>Index pipeline</summary>
<p>

```json
{
  "description": "drop old houses",
  "processors": [
    {
      "drop": {
        "if": "ZonedDateTime zdt = ZonedDateTime.parse(ctx.bazis_update_date); long millisDateTime = zdt.toInstant().toEpochMilli(); ZonedDateTime nowDate = ZonedDateTime.ofInstant(Instant.ofEpochMilli(millisDateTime), ZoneId.of('Z')); ZonedDateTime endDateZDT = ZonedDateTime.parse(ctx.end_date + 'T00:00:00Z'); long millisDateTimeEndDate = endDateZDT.toInstant().toEpochMilli(); ZonedDateTime endDate = ZonedDateTime.ofInstant(Instant.ofEpochMilli(millisDateTimeEndDate), ZoneId.of('Z')); return endDate.isBefore(nowDate);"
      }
    }
  ]
}
```

</p>
</details>


### version

Contains information about FIAS versions

<details><summary>Index mapping</summary>
<p>

```json
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
```

</p>
</details>


## Protobuf

#### Quick links

* [.proto files](app/interfaces/grpc/proto)

* [generated entities](app/infrastructure/persistence/grpc/dto)

* [grpc handlers](app/infrastructure/persistence/grpc/handler)

Use the code below if you want to recreate the generated entities.

#### Install

```shell script
mkdir tmp
cd tmp
git clone https://github.com/google/protobuf
cd protobuf
./autogen.sh
./configure
make
make check
sudo make install

cd $GOROOT/src/gofias/app
go get -u github.com/grpc-ecosystem/grpc-gateway/v1/protoc-gen-grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/v1/protoc-gen-swagger
go get -u github.com/golang/protobuf/protoc-gen-go
```

#### Generate proto
```shell script
export GOOGLEAPIS=$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.15.2/third_party/googleapis;\
protoc -I. -I$GOPATH/src -I$GOOGLEAPIS --go_out=plugins=grpc:. app/interfaces/grpc/proto/version/*.proto && \
protoc -I/usr/local/include -I. -I$GOOGLEAPIS --grpc-gateway_out=logtostderr=true:.  app/interfaces/grpc/proto/version/*.proto && \
protoc -I/usr/local/include -I. -I$GOOGLEAPIS --swagger_out=logtostderr=true:.  app/interfaces/grpc/proto/version/*.proto;\
protoc -I. -I$GOPATH/src -I$GOOGLEAPIS --go_out=plugins=grpc:. app/interfaces/grpc/proto/v1/address/*.proto && \
protoc -I/usr/local/include -I. -I$GOOGLEAPIS --grpc-gateway_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/address/*.proto && \
protoc -I/usr/local/include -I. -I$GOOGLEAPIS --swagger_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/address/*.proto;\
protoc -I. -I$GOPATH/src -I$GOOGLEAPIS --go_out=plugins=grpc:. app/interfaces/grpc/proto/health/*.proto && \
protoc -I/usr/local/include -I. -I$GOOGLEAPIS --grpc-gateway_out=logtostderr=true:.  app/interfaces/grpc/proto/health/*.proto && \
protoc -I/usr/local/include -I. -I$GOOGLEAPIS --swagger_out=logtostderr=true:.  app/interfaces/grpc/proto/health/*.proto;
```