package elasticsearch

import (
	"context"

	"go.uber.org/zap"
)

var indexName = "prom-metrics"

// const sampleType = "sample"
var activeIndexAlias = "active-"+indexName
var searchIndexAlias = "search-"+indexName
var activeIndexTemplate = `{
	"template": "active-`+indexName+`-*",
	"settings": {
		"number_of_shards":   5,
		"number_of_replicas": 1
	},
	"aliases": {
		"active-`+indexName+`":  {},
		"search-`+indexName+`": {}
	},
	"mappings":{
		"_default_": {
			"_all": {
				"enabled": false
			},
			"properties": {
				"timestamp": {
					"type": "date",
					"format": "strict_date_optional_time||epoch_millis"
				}
			},
			"dynamic_templates": [
				{
					"strings": {
						"match_mapping_type": "string",
						"path_match": "label.*",
						"mapping": {
							"type": "keyword"
						}
					}
				}
			]
		}
	}
}`

var inactiveIndexTemplate = `{
  "template": "inactive-`+indexName+`-*",
  "settings": {
	"number_of_shards":   1,
	"number_of_replicas": 0,
	"routing.allocation.include.box_type": "cold",
	"codec": "best_compression"
  }
}`

// ensureIndex creates the index in Elasticsearch.
func (a *Adapter) ensureIndex(ctx context.Context) error {
	_, err := a.c.IndexPutTemplate(activeIndexAlias).BodyString(activeIndexTemplate).Do(ctx)
	if err != nil {
		log.Fatal("Failed to create index template", zap.Error(err))
	}

	exists, err := a.c.IndexExists(activeIndexAlias).Do(ctx)
	if err != nil {
		return err
	}
	if !exists {
		a.c.CreateIndex(activeIndexAlias + "-000001").Do(ctx)
		if err != nil {
			log.Fatal("Failed to create initial index", zap.Error(err))
			return err
		}
	}
	return nil
}

// rolloverIndex
func (a *Adapter) rolloverIndex(ctx context.Context) error {
	_, err := a.c.RolloverIndex(activeIndexAlias).
		AddMaxIndexAgeCondition(a.indexMaxAge).
		AddMaxIndexDocsCondition(a.indexMaxDocs).
		Do(ctx)
	if err != nil {
		log.Error("Failed to rollover index", zap.Error(err))
		return err
	}
	return nil
}
