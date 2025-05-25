.PHONY: start-env
start-env:
	@docker compose up -d

.PHONY: stop-env
stop-env:
	@docker compose down

.PHONY: restart-env
restart-env: stop-env start-env

.PHONY: imports-fix
imports-fix:
	$(info $(M) fixing imports...)
	@go run go.xrstf.de/gimps@latest --config .gimps.yaml .