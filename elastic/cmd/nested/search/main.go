package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/Sereger/experiments/elastic/internal/index/nested"
	elastic "github.com/elastic/go-elasticsearch/v8"
)

var phrase = flag.String("q", "", "query")

func main() {
	flag.Parse()
	esClient, err := elastic.NewClient(elastic.Config{
		Addresses: []string{"http://localhost:9200", "http://localhost:9201"},
	})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	nestedIndex := nested.NewIndex(esClient)
	s := time.Now()
	docs, err := nestedIndex.Search(ctx, *phrase)
	d := time.Since(s)
	if err != nil {
		panic(err)
	}

	for _, doc := range docs {
		log.Printf("%d, %s\n", doc.ID, doc.Title)
		for _, child := range doc.Items {
			log.Printf("\t%d, %s\n", child.ID, child.Title)
		}
	}

	log.Println("done", d)
}
