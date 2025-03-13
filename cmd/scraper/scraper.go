package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/shajela/k8s-tool/internal/dbutils"
	"github.com/shajela/k8s-tool/internal/embeddings"
	"github.com/shajela/k8s-tool/internal/envutils"
	"github.com/shajela/k8s-tool/internal/schemas"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	ollama "github.com/ollama/ollama/api"
	"github.com/weaviate/weaviate/entities/models"

	"github.com/openai/openai-go"
)

const openAIBaseUrl = "https://api.openai.com/v1/embeddings"

type Spec struct {
	Namespace string
	Name      string
	Cpu       string
	Mem       string
}

type Pod struct {
	Spec      Spec
	Embedding []float32
}

func main() {
	scrapeInterval := 300
	if interval, err := envutils.LookupEnv("SCRAPE_INTERVAL"); err == nil {
		scrapeInterval, err = strconv.Atoi(interval)
		if err != nil {
			panic(fmt.Errorf("Error parsing SCRAPE_INTERVAL: %v", err))
		}
	}

	for {
		time.Sleep(time.Second * time.Duration(scrapeInterval))

		pm, err := poll()
		if err != nil {
			panic(err)
		}

		pods, err := embed(pm)
		if err != nil {
			panic(err)
		}

		err = push(pods)
		if err != nil {
			panic(err)
		}
	}
}

// Poll metrics server
func poll() (*v1beta1.PodMetricsList, error) {
	cfg, err := cfg()
	if err != nil {
		return nil, err
	}

	metricscs, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	pm, err := metricscs.MetricsV1beta1().PodMetricses("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pm, nil
}

// Determine whether to use local cfg
// or service account
func cfg() (*rest.Config, error) {
	// Ext means execution is happening
	// from outside of the cluster
	if _, err := envutils.LookupEnv("EXT"); err == nil {
		cname, err := envutils.LookupEnv("CLUSTER_NAME")
		if err != nil {
			return nil, err
		}

		cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{CurrentContext: cname},
		).ClientConfig()
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	// Else use service account
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// Parse relevant information and embed
func embed(pm *v1beta1.PodMetricsList) (*map[string]*Pod, error) {
	// Extract resource utilization from
	// each container
	pods := map[string]*Pod{}
	for _, i := range pm.Items {
		for _, c := range i.Containers {
			cpu := c.Usage[v1.ResourceCPU]
			mem := c.Usage[v1.ResourceMemory]

			pods[c.Name] = &Pod{
				Spec: Spec{
					Namespace: i.GetObjectMeta().GetNamespace(),
					Name:      i.GetObjectMeta().GetName(),
					Cpu:       cpu.String(),
					Mem:       mem.String(),
				},
			}
		}
	}

	// Whether using local model or OpenAI we
	// need to specify a model
	model, err := envutils.LookupEnv("EMBEDDING")
	if err != nil {
		return nil, err
	}

	// Embed locally during dev
	if dev, _ := envutils.LookupEnv("DEV"); dev == "true" {
		for _, p := range pods {
			body := ollama.EmbedRequest{
				Model: model,
				Input: fmt.Sprint(p.Spec),
			}

			payload, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}

			// Send request to ollama
			// assuming running in cluster
			ollamaBaseUrl := "http://host.docker.internal:11434/api/embed"
			if ext, _ := envutils.LookupEnv("EXT"); ext == "true" {
				// https://github.com/ollama/ollama/blob/main/docs/api.md
				ollamaBaseUrl = "http://localhost:11434/api/embed"
			}
			res, err := embeddings.ReqEmb(ollamaBaseUrl, payload, nil)
			if err != nil {
				return nil, fmt.Errorf("Error requesting embedding: %s", err.Error())
			}

			// Set cur pod's embedding value
			// to embedding from ollama
			emb := ollama.EmbedResponse{}
			err = json.Unmarshal(res, &emb)
			if err != nil {
				return nil, err
			}

			p.Embedding = emb.Embeddings[0]
		}
	} else {
		for _, p := range pods {
			body := map[string]string{
				"model": model,
				"input": fmt.Sprint(p.Spec),
			}

			payload, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}

			// Send request to OpenAI
			apiKey, err := envutils.LookupEnv("OPENAI_API_KEY")
			if err != nil {
				return nil, err
			}
			res, err := embeddings.ReqEmb(openAIBaseUrl, payload, map[string]string{
				"Content-Type":  "application/json",
				"Authorization": fmt.Sprintf("Bearer %s", apiKey),
			})
			if err != nil {
				return nil, err
			}

			emb := openai.CreateEmbeddingResponse{}
			err = json.Unmarshal(res, &emb)
			if err != nil {
				return nil, err
			}

			// TODO: Find a better way to do this
			t := make([]float32, len(emb.Data[0].Embedding))
			for i, _ := range emb.Data[0].Embedding {
				t[i] = float32(emb.Data[0].Embedding[i])
			}

			p.Embedding = t
		}
	}

	return &pods, nil
}

func push(pods *map[string]*Pod) error {
	// Get weaviate client
	client, err := dbutils.WeaviateClient()
	if err != nil {
		return err
	}

	// Check the connection
	rc, err := client.Misc().ReadyChecker().Do(context.Background())
	if err != nil || !rc {
		return fmt.Errorf("Could not establish connection to weaviate instance: %v\nReady checker: %t", err, rc)
	}

	moduleConfig := make(map[string]interface{})
	model, err := envutils.LookupEnv("MODEL")
	if err != nil {
		return err
	}

	// Use Ollama during dev
	if dev, _ := envutils.LookupEnv("DEV"); dev == "true" {
		// https://github.com/weaviate/weaviate/blob/main/modules/generative-ollama/config/class_settings.go
		moduleConfig["generative-ollama"] = map[string]interface{}{
			"apiEndpoint": "http://host.docker.internal:11434",
			"model":       model,
		}
	} else {
		moduleConfig["generative-openai"] = map[string]interface{}{
			"model": model,
		}
	}

	// Create schema for class if necessary
	class := models.Class{
		Class:        "Pod",
		Vectorizer:   "none",
		ModuleConfig: moduleConfig,
		InvertedIndexConfig: &models.InvertedIndexConfig{
			IndexTimestamps: true,
		},
		Properties: []*models.Property{
			{Name: "namespace", DataType: []string{"string"}},
			{Name: "name", DataType: []string{"string"}},
			{Name: "cpu", DataType: []string{"string"}},
			{Name: "mem", DataType: []string{"string"}},
		},
	}
	err = schemas.InitSchema(client, &class)

	if err != nil {
		return fmt.Errorf("Failed to create schema: %v", err)
	}
	o, err := class.MarshalBinary()
	if err != nil {
		return fmt.Errorf("Failed marshal class: %v", err)
	}
	log.Printf("Class:\n%s", string(o))

	// Create objects
	var objects []*models.Object
	for _, p := range *pods {
		objects = append(objects, &models.Object{
			Class: "Pod",
			Properties: map[string]any{
				"namespace": p.Spec.Namespace,
				"name":      p.Spec.Name,
				"cpu":       p.Spec.Cpu,
				"mem":       p.Spec.Mem,
			},
			Vector: p.Embedding,
		})
	}

	if err != nil {
		log.Fatalf("Failed to update schema: %v", err)
	}

	// Import
	batchRes, err := client.Batch().ObjectsBatcher().WithObjects(objects...).Do(context.Background())
	if err != nil {
		return err
	}

	for _, b := range batchRes {
		if b.Result.Errors != nil {
			errs := make([]string, len(b.Result.Errors.Error))
			for i, e := range b.Result.Errors.Error {
				errs[i] = e.Message
			}
			return fmt.Errorf("Batch res error: \n%s", strings.Join(errs, "\n"))
		}

		o, err := b.MarshalJSON()
		if err != nil {
			return err
		}
		log.Printf("%s\n%s", *b.Result.Status, string(o))
	}

	return nil
}
