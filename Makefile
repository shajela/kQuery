.PHONY: scrape query init start-db stop-db build deploy clean-weaviate run

init:
    ifneq (,$(wildcard ./.env))
        include .env
        export
    endif

start-db:
	@if [ -z "$$(docker ps -q -f name=weaviate-weaviate-1)" ]; then \
		echo "Database container is not running. Starting..."; \
		docker compose -f deployments/weaviate/compose.yaml up -d; \
	fi

stop-db:
	@if [ "$$(docker ps -q -f name=weaviate-weaviate-1)" ]; then \
		echo "Stopping database container..."; \
        docker compose -f deployments/weaviate/compose.yaml down; \
    fi

scraper: start-db init
	go run cmd/scraper/scraper.go

query: start-db init
	go run cmd/query/query.go

db-manager: start-db init
	go run cmd/db-manager/db-manager.go

build:
	docker build -t hajelasumer422/scraper:latest -f build/scraper/Dockerfile .
	docker build -t hajelasumer422/query:latest -f build/query/Dockerfile .
	docker build -t hajelasumer422/db-manager:latest -f build/db-manager/Dockerfile .

deploy:
	kubectl apply -f deployments/kubernetes/namespace.yaml
	kubectl apply -f deployments/kubernetes/rbac.yaml
	kubectl apply -f deployments/kubernetes/scraper.yaml -n kquery
	kubectl apply -f deployments/kubernetes/db-manager.yaml -n kquery
	kubectl apply -f deployments/kubernetes/query.yaml -n kquery
	kubectl apply -f deployments/kubernetes/service.yaml -n kquery

destroy:
	kubectl delete serviceaccount kquery -n kquery
	kubectl delete clusterrole kquery
	kubectl delete clusterrolebinding kquery
	kubectl delete deployments/scraper -n kquery
	kubectl delete deployments/db-manager -n kquery
	kubectl delete deployments/query -n kquery
	kubectl delete service query-svc -n kquery
	kubectl delete namespace kquery

clean-weaviate:
	docker volume rm weaviate_weaviate_data

run:
	kubectl port-forward service/query-svc 30010:$(PORT) -n kquery
