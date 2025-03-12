package dbutils

import (
	"fmt"
	"strings"

	"github.com/shajela/k8s-tool/internal/envutils"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

func WeaviateClient() (*weaviate.Client, error) {
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
		return nil, fmt.Errorf("Could not create weaviate client: %v", err)
	}

	return client, nil
}

func HandleErr(res *models.GraphQLResponse, err error) error {
	if err != nil {
		return err
	} else if res.Errors != nil {
		errs := make([]string, len(res.Errors))
		for i, e := range res.Errors {
			errs[i] = e.Message
		}
		return fmt.Errorf("Error in GraphQL response:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}
