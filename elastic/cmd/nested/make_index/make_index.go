package main

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/Sereger/experiments/elastic/internal/index/nested"
	elastic "github.com/elastic/go-elasticsearch/v8"
)

func main() {
	esClient, err := elastic.NewClient(elastic.Config{
		Addresses: []string{"http://localhost:9200", "http://localhost:9201"},
	})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	nestedIndex := nested.NewIndex(esClient)
	err = nestedIndex.CreateIndex(ctx, true)
	if err != nil {
		panic(err)
	}

	gen := &generator{rnd: rand.New(rand.NewSource(time.Now().Unix()))}
	err = nestedIndex.Add(ctx, gen.newDocs(1000, 100))
	if err != nil {
		panic(err)
	}

	log.Println("done")
}

type generator struct {
	rnd *rand.Rand
	seq int
}

func (g *generator) newDocs(n int, children int) []*nested.Doc {
	docs := make([]*nested.Doc, n)
	for i := 0; i < n; i++ {
		docs[i] = g.newDoc(children)
	}

	return docs
}

func (g *generator) newDoc(children int) *nested.Doc {
	g.seq++
	doc := &nested.Doc{ID: g.seq}
	words := g.rnd.Intn(3) + 1
	builder := strings.Builder{}
	for i := 0; i < words; i++ {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(dict[g.rnd.Intn(len(dict))])
	}
	doc.Title = builder.String()

	if children > 0 {
		doc.Items = g.newDocs(children, 0)
	}

	return doc
}

var dict = []string{
	"рыба", "борщ", "харчо", "пицца", "сырный суп", "пельмени", "спагетти", "хачапури", "бифштекс",
	"блины", "бургер", "лазанья", "шашлык", "индейка", "цезарь", "оливки", "круасан", "икра",
	"вкусный", "сытный", "пряный", "кислый", "свежий",
}
