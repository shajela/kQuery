.PHONY: scrape query init start-db stop-db build deploy

init:
    ifneq (,$(wildcard ./.env))
        include .env
        export
    endif

start-db:
	@if [ -z "$$(docker ps -q -f name=deployments-weaviate-1)" ]; then \
		echo "Database container is not running. Starting..."; \
		docker compose -f deployments/weaviate/compose.yaml up -d; \
	fi

stop-db:
	@if [ "$$(docker ps -q -f name=deployments-weaviate-1)" ]; then \
		echo "Stopping database container..."; \
        docker compose -f deployments/weaviate/compose.yaml down; \
    fi

scraper: start-db init
	go run cmd/scraper/scraper.go

query: start-db init
	go run cmd/query/query.go

build:
	docker build -t hajelasumer422/scraper:latest -f build/scraper/Dockerfile .
	docker build -t hajelasumer422/query:latest -f build/query/Dockerfile .

deploy:
	kubectl create -f deployments/kubernetes/namespace.yaml
	kubectl create -f deployments/kubernetes/rbac.yaml
	kubectl apply -f deployments/kubernetes/scraper.yaml -n kquery

destroy:
	kubectl delete serviceaccount kquery
	kubectl delete clusterrole kquery
	kubectl delete clusterrolebinding kquery
	kubectl delete deployments/scraper
	kubectl delete namespace kquery
