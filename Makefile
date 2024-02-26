.PHONY: migrate
migrate:
	docker compose -f compose.base.yml run --rm migrate

.PHONY: sql
sql:
	docker compose -f compose.base.yml run --rm sql

.PHONY: test
test:
	docker compose --env-file ./.env.testing -f compose.test.yml run --rm test
