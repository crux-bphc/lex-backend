.PHONY: migrate
migrate:
	docker compose --profile debug run --rm migrate

.PHONY: sql
sql:
	docker compose --profile debug run --rm sql

.PHONY: test
test:
	docker compose --env-file ./.env.testing -f compose.test.yml run --rm test
