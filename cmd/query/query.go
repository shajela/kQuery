package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shajela/k8s-tool/internal/dbutils"
	"github.com/shajela/k8s-tool/internal/envutils"
	"github.com/shajela/k8s-tool/internal/generate"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

// https://github.com/ollama/ollama/blob/main/docs/api.md
const ollamaBasePort = "11434"
const ollamaBaseUrl = "http://localhost:11434/api/embed"

const openAIBaseUrl = "https://api.openai.com/v1/embeddings"

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter query: ")
	query, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	query = strings.TrimSpace(query)

	res, err := genRes(query)
	if err != nil {
		panic(err)
	}
	log.Println("Generated response:", res)

	mb, err := res.MarshalBinary()
	if err != nil {
		panic(err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(mb, &data)
	if err != nil {
		panic(err)
	}
	log.Println("Extracted response:", *generate.ExtractGroupedResult(data))
}

func genRes(query string) (*models.GraphQLResponse, error) {
	weaviteHost, err := envutils.LookupEnv("WEAVITE_HOST")
	if err != nil {
		return nil, err
	}

	cfg := weaviate.Config{
		Host:   weaviteHost,
		Scheme: "http",
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	gs := graphql.NewGenerativeSearch().GroupedResult(fmt.Sprintf("Answer the following query. If needed, you can count or perform any mathematical operations. If you don't know the answer, simply say that you don't know. Keep answers concise. The query is: %s", query))

	res, err := client.GraphQL().Get().
		WithClassName("Pod").
		WithFields(
			graphql.Field{Name: "name"},
			graphql.Field{Name: "namespace"},
			graphql.Field{Name: "cpu"},
			graphql.Field{Name: "mem"},
		).
		WithGenerativeSearch(gs).
		Do(context.Background())

	err = dbutils.HandleErr(res, err)
	if err != nil {
		return nil, err
	}

	return res, nil
}
