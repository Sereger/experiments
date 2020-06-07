package nested

const query = `{
  "_source": {
    "includes": ["*"],
    "excludes": ["items"]
  },
  "min_score": 0.0001,
  "query": {
    "bool": {
      "should": [
        {
          "nested": {
            "path": "items",
            "inner_hits": {
              "_source": [
                "items.id", "items.title"
              ]
            },
            "query": {
              "bool": {
                "must": {
                  "multi_match": {
                    "query": "%s",
                    "fields": [
                      "items.title"
                    ],
                    "operator": "or",
                    "type": "cross_fields"
                  }
                }
              }
            }
          }
        },
        {
          "multi_match": {
            "query": "%s",
            "fields": [
              "title^2"
            ],
            "operator": "or",
            "type": "cross_fields"
          }
        }
      ]
    }
  },
	"sort" : [
		{ "_score" : "desc" }
	],
	"size" : 500,
  	"from" : 0
}`
