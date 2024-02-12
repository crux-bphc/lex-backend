ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: migrate
migrate:
	docker compose -f compose.base.yml run --rm migrate

.PHONY: cron
cron:
	docker compose -f compose.base.yml run --rm -v $(ROOT_DIR)/cron:/app cron python $(task).py