# GoFias

[Go](http://www.golang.org/) (golang) библиотека для импорта данных из [БД ФИАС](https://fias.nalog.ru/) в [Elasticsearch](http://www.elasticsearch.org/)

Прочитать на других языках: [Русский](README.md), [English](README.en.md).

## Базовые параметры командной строки
* `config-path (строка)` - Установить путь конфигурации (по умолчанию `./`)
* `config-type (строка)` - Установите тип конфигурации (по умолчани `yaml`)
* `logger-prefix (строка)` - Префикс директории для логирования (по умолчани `cli`)

## Использование сервиса импорта ФИАС
```shell script
cd GOROOT/src/gofias/app/
go build -o ./fias ./application/cli/
cd ..
./fias update --skip-houses --skip-clear
```

## Параметры командной строки сервиса импорта ФИАС
* `skip-clear (булево)` - Пропустить очистку каталога при запуске (по умолчанию `false`)
* `skip-houses (булево)` - Пропустить импорт домов (default `false`)
* `skip-osm (булево)` - Пропустить импорт гео-данных (default `false`)

## Использование GRPC-сервера

### С использованием docker (docker-compose)
```shell script
cd GOROOT/src/gofias
docker-compose up -d
```

### Без использования docker (docker-compose)
```shell script
cd GOROOT/src/gofias
go build -o ./fias ./application/grpc/
cd ..
./fias --logger-prefix=grpc grpc
```

## Информация об индексах ElasticSearch

### Адреса (address)

Содержит информацию об адресах ФИАС

<details><summary>Структура индекса</summary>
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

<details><summary>Обработчик удаления старых адресов</summary>
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

### Дома (houses)

Содержит информацию о домах ФИАС

<details><summary>Структура индекса</summary>
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

<details><summary>Обработчик удаления старых домов</summary>
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

Содержит информацию о версиях ФИАС

<details><summary>Структура индекса</summary>
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


## Формат Protobuf

#### Быстрые ссылки

* [.proto-файлы](app/interfaces/grpc/proto)

* [сгенерированные сущности](app/infrastructure/persistence/grpc/dto)

* [grpc-обоработчкики запросов](app/infrastructure/persistence/grpc/handler)

Используйте приведенный ниже код, если Вы хотите перегенерировать созданные ранее сущности.

#### Установка protoc

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
go get -u github.com/securego/gosec/v2/cmd/gosec
```

#### Генерация сущностей
```shell script
export SWAGGER_OPTIONS=$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.15.2;\
export GOOGLEAPIS=SWAGGER_OPTIONS/third_party/googleapis;\
protoc -I. -I$GOPATH/src -I$GOOGLEAPIS -I$SWAGGER_OPTIONS --go_out=plugins=grpc:. app/interfaces/grpc/proto/v1/fias/*.proto && \
protoc -I/usr/local/include -I. -I$GOOGLEAPIS -I$SWAGGER_OPTIONS --grpc-gateway_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/fias/*.proto && \
protoc -I/usr/local/include -I. -I$GOOGLEAPIS -I$SWAGGER_OPTIONS --swagger_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/fias/*.proto;
```