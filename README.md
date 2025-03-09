# kQuery

kQuery is a CLI tool that retrieves kubernetes metric data from natural language queries. Please note that this project is under active development and is currently just in a proof-of-concept stage.

## Requirements
- [Kubernetes Metrics Server](https://github.com/kubernetes-sigs/metrics-server): Used to embed kubernetes component metrics
- [Ollama](https://ollama.com/): Used to run models locally
- [OpenAI](https://openai.com/api/): API key used to run models
- [Go](https://go.dev/doc/install): Go `v1.24.0` is preferred
- [Docker](https://docs.docker.com/): Current architecture deploys a containerized [Weaviate](https://weaviate.io/) instance on the local machine for vector storage

## Environment Variables
| Name      | Description | Default | Required
| --------- | ----------- | ------- | --------
| `EXT`  | If the application is being run locally, set to 'true.' If the application is deployed in the cluster itself, set to 'false.' Note: the latter behavior is not yet supported. | false | No |
| `CLUSTER_NAME` | The cluster to be monitored. Required if running application code outside of the cluster. | - | No |
| `DEV`    | To use the Ollama API, set to 'true.' To use the OpenAI API, set to 'false.' | false | No |
| `EMBEDDING` | The embedding model to be used. For more information, check the documentation of the configured provider. | - | Yes 
| `MODEL` | The generative model to be used. For more information, check the documentation of the configured provider. | - | Yes 
| `WEAVITE_HOST` | The address of the Weaviate instance to connect to. | - | Yes

## Usage
To populate Weaviate instance:
```
make scrape
```

To make a query:
```
make query
```

To start Weaviate instance (note: instance will automatically start if `make query` or `make scrape` is run and no instance is detected):
```
make start-db
```

To stop Weaviate instance:
```
make stop-db
```

## Example Use-case
```
make query
go run cmd/query/query.go
Enter query: What is the average CPU usage of pods in the kube-system namespace?

To calculate this, I'll sum up the CPU usages and divide by the number of pods:

Sum of CPU usages:
37763196n + 554038n + 17727950n + 92825n + 11262659n + 2729837n + 4138960n + 20554174n + 605416n = 43231131n

Number of pods: 9

Average CPU usage:
43231131n ÷ 9 = 4803396.77n (approximately)

Answer: The average CPU usage is approximately 4,803,396.77 nanocores.
```