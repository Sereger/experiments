package cmd

import (
	"context"
	"github.com/Sereger/experiments/elastic/internal/index/flat"
	"github.com/Sereger/experiments/elastic/internal/index/nested"
	elastic "github.com/elastic/go-elasticsearch/v8"
	"testing"
)

func BenchmarkIndexes(b *testing.B) {
	esClient, err := elastic.NewClient(elastic.Config{
		Addresses: []string{"http://localhost:9200", "http://localhost:9201"},
	})
	if err != nil {
		b.FailNow()
	}

	b.Run("nested", func(b *testing.B) {
		nestedIndex := nested.NewIndex(esClient)
		b.RunParallel(func(pb *testing.PB) {
			var i int
			for pb.Next() {
				phrase := dict[i%len(dict)]
				_, err := nestedIndex.Search(context.Background(), phrase)
				if err != nil {
					b.Error(err)
				}
				i++
			}
		})
	})

	b.Run("flat", func(b *testing.B) {
		flatIndex := flat.NewIndex(esClient)
		b.RunParallel(func(pb *testing.PB) {
			var i int
			for pb.Next() {
				phrase := dict[i%len(dict)]
				_, err := flatIndex.Search(context.Background(), phrase)
				if err != nil {
					b.Error(err)
				}
				i++
			}
		})
	})
}

var dict = []string{
	"рыба", "борщ", "харчо", "пицца", "сырный суп", "пельмени", "спагетти", "хачапури", "бифштекс",
	"блины", "бургер", "лазанья", "шашлык", "индейка", "цезарь", "оливки", "круасан", "икра",
	"вкусный", "сытный", "пряный", "кислый", "свежий",
}
