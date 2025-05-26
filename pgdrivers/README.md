Пример окружения, где можно воспроизвести проблему с prepared statements при работе в стеке go + pgbouncer + postgresql.
Доступны след. make-команды:
```shell
make start-env         # запуск окружения с использованием docker compose
make stop-env          # остановка окружения
make start-env-with-ps # запуск окружения с настройкой MAX_PREPARED_STATEMENTS=200
```

`main.go` можно запускать со след параметрами:
- mode:
-- `lib/pq` - запуск теста с драйвером lib/pq
-- `jackc/pgx` - запуск теста с драйвером jackc/pgx
-- `exec` - запуск теста с драйвером jackc/pgx c настройкой `default_query_exec_mode=exec`

- problem: признак, сигнализирующий о том, что тест будет запущен в режиме демонстрации проблемы (не совместим с `mode=exec`).
- n: количество запросов, которое будет выполнено в ходе теста (по умолчанию 10) 
Пример запуска `main.go`:
```shell
go run main.go --mode lib/pq --problem
```
Пример выше запустит тест с драйвером lib/pq в режиме демонстрации проблемы.

