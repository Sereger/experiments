package nested

const (
	analyzer = `{
    "number_of_shards": 1,
    "number_of_replicas": 1,
    "analysis": {
      "filter": {
        "russian_stop": {
          "type": "stop",
          "stopwords": "_russian_"
        },
        "russian_stemmer": {
          "type": "stemmer",
          "language": "russian"
        },
        "russian_phonetic": {
          "type": "russian_phonetic",
          "replace": false,
          "vowels": "encode_all"
        },
        "synonym_search": {
          "type": "synonym",
          "synonyms": [
            "шаверм, шаурм",
            "суши, сашими, ролл",
            "рыб, рыбн",
            "иль патио, ильпатио, il patio, il патио => ilpatio",
            "papa john’s, papa johns, papajohns, пап джонс => papa_johns",
            "альп гольд, альп голд, alpen gold => alpen_gold",
            "red bull, рэд бул => red_bull",
            "red machine, рэд машин => red_machine"
          ]
        },
        "synonym_all": {
          "type": "synonym",
          "synonyms": [
            "тунц => тунец, рыб",
            "семг => семг, рыб",
            "горбуш => горбуш, рыб",
            "лос => лос, рыб",
            "форел => форел, рыб",
            "харч => харч, суп",
            "лагма => лагма, суп",
            "борщ => борщ, суп",
            "окрошк => окрошк, суп",
            "солянк => солянк, суп"
          ]
        },
        "synonym_name": {
          "type": "synonym",
          "synonyms": [
            "иль патио, ильпатио, il patio, il патио => ilpatio",
            "papa john’s, papa johns, papajohns, пап джонс => papa_johns",
            "альп гольд, альп голд, alpen gold => alpen_gold",
            "red bull, рэд бул => red_bull"
          ]
        }
      },
      "analyzer": {
        "witch_russian": {
          "tokenizer": "standard",
          "filter": [
            "lowercase",
            "russian_stop",
            "russian_stemmer",
            "synonym_all",
            "russian_phonetic"
          ]
        },
        "witch_russian_search": {
          "tokenizer": "standard",
          "filter": [
            "lowercase",
            "russian_stop",
            "russian_stemmer",
            "synonym_search",
            "russian_phonetic"
          ]
        }
      }
    }
  }`
	mapping = `{
    "properties": {
      "id": {
        "type": "keyword"
      },
      "title": {
        "type": "text",
        "store": true,
        "fielddata": true,
        "analyzer": "witch_russian",
        "search_analyzer": "witch_russian_search"
      },
      "items": {
        "type": "nested",
        "properties": {
          "id": {
            "type": "keyword"
          },
          "title": {
            "type": "text",
            "store": true,
            "fielddata": true,
            "analyzer": "witch_russian",
            "search_analyzer": "witch_russian_search"
          }
        }
      }
    }
  }`
	config    = `{"settings": ` + analyzer + `,"mappings": ` + mapping + `}`
	indexName = "nested_v1"
)
