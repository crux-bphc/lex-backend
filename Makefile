.PHONY: migrate
migrate:
	docker compose -f compose.base.yml run --rm migrate

.PHONY: sql
sql:
	docker compose -f compose.base.yml run --rm sql