.PHONY: start-env
start-env:
	@MAX_PREPARED_STATEMENTS=0 docker compose up -d

.PHONY: start-env-with-ps
start-env-with-ps:
	@MAX_PREPARED_STATEMENTS=200 docker compose up -d

.PHONY: stop-env
stop-env:
	@docker compose down

.PHONY: restart-env
restart-env: stop-env start-env

.PHONY: imports-fix
imports-fix:
	$(info $(M) fixing imports...)
	@go run go.xrstf.de/gimps@latest --config .gimps.yaml .

