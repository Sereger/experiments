build:
	docker build -t my-kafka kafka-docker/
up:
	docker run --rm -d --name my-kafka -p 2181:2181 -p 9092:9092 -e ADVERTISED_PORT=9092 -e ADVERTISED_HOST=127.0.0.1 my-kafka

down:
	docker stop my-kafka

attach:
	docker run --rm -it --name my-kafka -p 2181:2181 -p 9092:9092 my-kafka bash