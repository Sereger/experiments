# Бэнчмарк для поиска вложенных документов (nested) в ElasticSearch
Наполнение индексов (создвется 100 000 документов):
```shell script
go run cmd/nested/make_index/make_index.go  # nested index
go run cmd/flat/make_index/make_index.go  # flat index
```

Выполнение запросов:
```shell script
go run cmd/nested/search/main.go -q шашлык
go run cmd/flat/search/main.go -q шашлык
```

Запуск бэнчмарка:
```shell script
go test -bench . -benchmem -v ./cmd/bench_test.go
```

Результат для MacBook Pro (13-inch, 2019, 2,8 GHz Intel Core i7, 16 ГБ 2133 MHz LPDDR3):
```shell script
goos: darwin
goarch: amd64
BenchmarkIndexes/nested-8                     10         123739892 ns/op         1487539 B/op      15331 allocs/op
BenchmarkIndexes/flat-8                       91          12602806 ns/op          298909 B/op       3160 allocs/op
```