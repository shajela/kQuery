# kQuery

kQuery is a Kubernetes-native framework that retrieves metric data from natural language queries. Driven by RAG architecture, kQuery runs within the cluster to provide seamless access to real-time insights.

![demo](https://github.com/shajela/kQuery/blob/main/demo.gif)

## Requirements
- [Kubernetes Metrics Server](https://github.com/kubernetes-sigs/metrics-server): Used to embed kubernetes component metrics
- [Ollama](https://ollama.com/): Used to serve models locally
- [OpenAI](https://openai.com/api/): API key used to run models
- [Go](https://go.dev/doc/install): Go `v1.24.0` is preferred (if running locally)
- [Docker](https://docs.docker.com/): Current architecture deploys a containerized [Weaviate](https://weaviate.io/) instance on the local machine for vector storage

## Environment Variables
| Name      | Pod | Description | Default | Required
| --------- | -------- | ----------- | ------- | --------
| `EXT`  | - | If the application is being run locally, set to 'true.' If the application is deployed in the cluster itself, set to 'false.' | false | No |
| `CLUSTER_NAME` | - | The cluster to be monitored. Required if running application code outside of the cluster. | - | No |
| `DEV`    | `scraper` | To use the Ollama API, set to 'true.' To use the OpenAI API, set to 'false.' If false, `OPEN_AI_API_KEY` must be provided. | false | No |
| `OPENAI_API_KEY`    | `scraper` | OpenAI API key. Must be set if `DEV` is set to false. Additionally, `OPENAI_APIKEY` environment variable must be configured in the Weaviate instance. See [Weaviate docs](https://weaviate.io/developers/weaviate/model-providers/openai/generative) for more information. | - | No |
| `EMBEDDING` | `scraper` | The embedding model to be used. For more information, check the documentation of the configured provider. | - | Yes 
| `MODEL` | `scraper` | The generative model to be used. For more information, check the documentation of the configured provider. | - | Yes 
| `WEAVITE_HOST` | `query`,`scraper`,`db-manager` | The address of the Weaviate instance to connect to. | - | Yes
| `CLEANUP_INTERVAL` | `db-manager` | The amount of time in seconds to wait between cleanup operations for the Weaviate instance. | 300 | No
| `RETENTION_PERIOD` | `db-manager` | During a cleanup operation, the application will take the current time and subtract `RETENTION_PERIOD`. Any entries older than this calculated timestamp will be deleted from the Weaviate instance. Configured in seconds. | 86400 | No
| `SCRAPE_INTERVAL` | `db-manager` | The number of seconds to wait between consecutive exectutions of scraping the metrics API, embedding, and pushing to the Weaviate instance. | 300 | No
| `PORT` | `query` | The port of `query` to serve requests. | - | Yes

## Installation
Below is an example of how to deploy `kQuery` in a running kubernetes cluster. For the example, we will be using [these](https://github.com/shajela/kQuery/tree/main/deployments/kubernetes) example manifests and [this](https://github.com/shajela/kQuery/tree/main/deployments/weaviate) config. The cluster in this example is running using [kind](https://kind.sigs.k8s.io/) and all deployment steps are automated using [this](https://github.com/shajela/kQuery/blob/main/Makefile) Makefile.

1. Deploy weaviate instance and all kubernetes resources with `make start-all`
2. Open the connection between the local machine and the kind cluster with `make run`
3. Send queries to kQuery by running `client.go`
