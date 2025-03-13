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

	// Keep for one day by default
	rp := 60 * 60 * 24
	if seconds, err := envutils.LookupEnv("RETENTION_PERIOD"); err == nil {
		rp, err = strconv.Atoi(seconds)
		if err != nil {
			panic(fmt.Errorf("Error parsing RETENTION_PERIOD: %v", err))
		}
	}

	where := filters.Where().
		WithPath([]string{"_creationTimeUnix"}).
		WithOperator(filters.LessThan).
		WithValueDate(time.Now().Add(-time.Second * time.Duration(rp)))

	res, err := client.Batch().ObjectsBatchDeleter().
		WithClassName("Pod").
		WithWhere(where).
		WithOutput("verbose").
		Do(context.Background())
	if err != nil {
		return err
	}

	for _, o := range res.Results.Objects {
		log.Printf("Deleted object with id %v", o.ID)
	}
	return nil
}
