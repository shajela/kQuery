package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/shajela/k8s-tool/internal/dbutils"
	"github.com/shajela/k8s-tool/internal/envutils"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

func main() {
	cleanupInterval := 300
	if interval, err := envutils.LookupEnv("CLEANUP_INTERVAL"); err == nil {
		cleanupInterval, err = strconv.Atoi(interval)
		if err != nil {
			panic(fmt.Errorf("Error parsing CLEANUP_INTERVAL: %v", err))
		}
	}

	for {
		time.Sleep(time.Second * time.Duration(cleanupInterval))
		err := cleanup()
		if err != nil {
			panic(err)
		}
	}
}

// Cleanup old objects from DB
// https://github.com/weaviate/weaviate/pull/2376
func cleanup() error {
	client, err := dbutils.WeaviateClient()
	if err != nil {
		return err
	}

	// TODO: delete n oldest objects
	where := filters.Where().
		WithPath([]string{"_creationTimeUnix"}).
		WithOperator(filters.LessThan).
		WithValueDate(time.Now())
	res, err := client.GraphQL().Get().
		WithClassName("Pod").
		WithFields(
			graphql.Field{Name: "name"},
		).
		WithWhere(where).
		Do(context.Background())

	err = dbutils.HandleErr(res, err)
	if err != nil {
		return err
	}
	log.Println(res)
	return nil
}
