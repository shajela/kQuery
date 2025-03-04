.PHONY: scrape query init start-db stop-db

init:
    ifneq (,$(wildcard ./.env))
        include .env
        export
    endif

start-db:
	@if [ -z "$$(docker ps -q -f name=deployments-weaviate-1)" ]; then \
		echo "Database container is not running. Starting..."; \
		docker compose -f deployments/compose.yaml up -d; \
	fi

stop-db:
	@if [ "$$(docker ps -q -f name=deployments-weaviate-1)" ]; then \
		echo "Stopping database container..."; \
        docker compose -f deployments/compose.yaml down; \
    fi

scrape: start-db init
	go run cmd/scraper/scraper.go

query: start-db init
	go run cmd/query/query.go
