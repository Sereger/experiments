package flat

const query = `{
  "min_score": 0.0001,
  "query": {
    "bool": {
      "must": {
		  "multi_match": {
			"query": "%s",
			"fields": [
			  "title"
			],
			"operator": "or",
			"type": "cross_fields"
		  }
		}
    }
  },
	"sort" : [
		{ "_score" : "desc" }
	],
	"size" : 500,
  	"from" : 0
}`
