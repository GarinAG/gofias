# GoFias

gofias is a [Go](http://www.golang.org/) (golang) library that import [fias](https://fias.nalog.ru/) data to [Elasticsearch](http://www.elasticsearch.org/)

## Usage

```shell script
git clone https://github.com/GarinAG/gofias.git
cd gofias
go build ./src/gofias/
./gofias --host=localhost:9200 --bulk-size=5000 --status --skip-houses --skip-snapshot --skip-clear
```

## CLI props

* `bulk-size (int)` - Number of documents to collect before committing (default `1000`)
* `cpu (int)` - Count of CPU usage (default `0`)
* `force (bool)` - Force full download (default `false`)
* `force-index (bool)` - Start force index without import (default `false`)
* `host (string)` - Elasticsearch host (default `"localhost:9200"`)
* `logs (string)` - Logs dir path (default `"./logs"`)
* `prefix (string)` - Prefix for elasticsearch indexes (default `"fias_"`)
* `skip-clear (bool)` - Skip clear tmp directory on start (default `false`)
* `skip-houses (bool)` - Skip houses index (default `false`)
* `skip-snapshot (bool)` - Skip create ElasticSearch snapshot (default `false`)
* `status (bool)` - Show import status (default `false`)
* `storage (string)` - Snapshots storage path (default `"/usr/share/elasticsearch/snapshots"`)
* `tmp (string)` - Tmp folder relative path in user home dir (default `"/tmp/fias/"`)


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
      "number_of_replicas": "0",
      "refresh_interval": "-1",
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
          "autocomplete_filter": {
            "type": "edge_ngram",
            "min_gram": 2,
            "max_gram": 20
          },
          "fias_word_delimiter": {
            "type": "word_delimiter",
            "preserve_original": "true",
            "generate_word_parts": "false"
          }
        },
        "analyzer": {
          "autocomplete": {
            "type": "custom",
            "tokenizer": "standard",
            "filter": ["autocomplete_filter"]
          },
          "stop_analyzer": {
            "type": "custom",
            "tokenizer": "whitespace",
            "filter": ["lowercase", "fias_word_delimiter"]
          }
        }
      }
    }
  },
  "mappings": {
    "dynamic": false,
    "properties": {
      "street_address_suggest": {
        "type": "text",
        "analyzer": "autocomplete",
        "search_analyzer": "stop_analyzer"
      },
      "full_address": {
        "type": "keyword"
      },
      "district_full": {
        "type": "keyword"
      },
      "settlement_full": {
        "type": "keyword"
      },
      "street_full": {
        "type": "keyword"
      },
      "formal_name": {
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
      "oper_status": {
        "type": "integer"
      },
      "act_status": {
        "type": "integer"
      },
      "live_status": {
        "type": "integer"
      },
      "cent_status": {
        "type": "integer"
      },
      "ao_guid": {
        "type": "keyword"
      },
      "parent_guid": {
        "type": "keyword"
      },
      "ao_level": {
        "type": "keyword"
      },
      "area_code": {
        "type": "keyword"
      },
      "auto_code": {
        "type": "keyword"
      },
      "city_ar_code": {
        "type": "keyword"
      },
      "city_code": {
        "type": "keyword"
      },
      "street_code": {
        "type": "keyword"
      },
      "extr_code": {
        "type": "keyword"
      },
      "sub_ext_code": {
        "type": "keyword"
      },
      "place_code": {
        "type": "keyword"
      },
      "plan_code": {
        "type": "keyword"
      },
      "plain_code": {
        "type": "keyword"
      },
      "code": {
        "type": "keyword"
      },
      "postal_code": {
        "type": "keyword"
      },
      "region_code": {
        "type": "keyword"
      },
      "street": {
        "type": "keyword"
      },
      "district": {
        "type": "keyword"
      },
      "district_type": {
        "type": "keyword"
      },
      "street_type": {
        "type": "keyword"
      },
      "settlement": {
        "type": "keyword"
      },
      "settlement_type": {
        "type": "keyword"
      },
      "okato": {
        "type": "keyword"
      },
      "oktmo": {
        "type": "keyword"
      },
      "ifns_fl": {
        "type": "keyword"
      },
      "ifns_ul": {
        "type": "keyword"
      },
      "terr_ifns_fl": {
        "type": "keyword"
      },
      "terr_ifns_ul": {
        "type": "keyword"
      },
      "norm_doc": {
        "type": "keyword"
      },
      "start_date": {
        "type": "date"
      },
      "end_date": {
        "type": "date"
      },
      "bazis_finish_date": {
        "type": "date"
      },
      "bazis_create_date": {
        "type": "date"
      },
      "bazis_update_date": {
        "type": "date"
      },
      "update_date": {
        "type": "date"
      },
      "location": {
        "type": "geo_point"
      },
      "houses": {
        "type": "nested",
        "properties": {
          "houseId": {
            "type": "keyword"
          },
          "build_num": {
            "type": "keyword"
          },
          "house_num": {
            "type": "keyword"
          },
          "str_num": {
            "type": "keyword"
          },
          "ifns_fl": {
            "type": "keyword"
          },
          "ifns_ul": {
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
          "update_date": {
            "type": "date"
          },
          "cad_num": {
            "type": "keyword"
          },
          "terr_ifns_fl": {
            "type": "keyword"
          },
          "terr_ifns_ul": {
            "type": "keyword"
          },
          "okato": {
            "type": "keyword"
          },
          "oktmo": {
            "type": "keyword"
          },
          "location": {
            "type": "geo_point"
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
      "if": "ctx.curr_status  != '0' "
    }
  }, {
    "drop": {
      "if": "ctx.act_status  != '1'"
    }
  }, {
    "drop": {
      "if": "ctx.live_status  != '1'"
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
      "number_of_replicas": "0",
      "refresh_interval": "-1",
      "requests": {
        "cache": {
          "enable": "true"
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
      "ao_guid": {
        "type": "keyword"
      },
      "build_num": {
        "type": "keyword"
      },
      "house_num": {
        "type": "keyword"
      },
      "str_num": {
        "type": "keyword"
      },
      "ifns_fl": {
        "type": "keyword"
      },
      "ifns_ul": {
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
      "bazis_finish_date": {
        "type": "date"
      },
      "bazis_create_date": {
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
      "terr_ifns_fl": {
        "type": "keyword"
      },
      "terr_ifns_ul": {
        "type": "keyword"
      },
      "okato": {
        "type": "keyword"
      },
      "oktmo": {
        "type": "keyword"
      },
      "location": {
        "type": "geo_point"
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


### info

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
```

</p>
</details>