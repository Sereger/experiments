chRun:
	docker run -it --rm -P --name clickhouse-server --ulimit nofile=262144:262144 yandex/clickhouse-server

tRun:
	docker run --rm -t -i -P --name tarantool-server -p3301:3301 -e TARANTOOL_USER_NAME=u -e TARANTOOL_USER_PASSWORD=tarPass tarantool/tarantool

mongoRun:
	docker run --rm -p27017:27017 mongo
data:
	@echo "get data from resource http://services.fms.gov.ru/info-service.htm?sid=2000"
