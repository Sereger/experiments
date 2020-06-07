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