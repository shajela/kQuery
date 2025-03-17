package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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

type Request struct {
	Body string
}

func main() {
	port, err := envutils.LookupEnv("PORT")
	if err != nil {
		panic(err)
	}
	serve(port)
}

func serve(port string) {
	log.Println("Starting controller on", port)
	http.HandleFunc("/", receiveReq)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

// Recieve request from client
func receiveReq(w http.ResponseWriter, r *http.Request) {
	log.Printf("Recieved request %+v", r)

	if r.URL.Path != "/" {
		log.Printf("Recieved unknown path '%s'", r.URL.Path)
		return
	}

	// Check header
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	// Limit maximum read of 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req Request
	err := dec.Decode(&req)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError

		switch {
		// Syntax error
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Syntax error: https://github.com/golang/go/issues/25956
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

		// Type error
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Type error: https://github.com/golang/go/issues/29035
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// Empty body
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// Request too large
		case errors.As(err, &maxBytesError):
			msg := fmt.Sprintf("Request body must not be larger than %d bytes", maxBytesError.Limit)
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// Else send 500 error
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// Check only recieved a single JSON object
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	log.Printf("Decoded request %+v", req)
	res, err := genRes(req.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	log.Println("Generated response:", res)
	mb, err := res.MarshalBinary()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(mb, &data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(*generate.ExtractGroupedResult(data)))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println("Error writing HTTP reply: %v", err)
		return
	}
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
