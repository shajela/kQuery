package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nEnter query ('exit' to leave): ")
		query, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		query = strings.TrimSpace(query)

		if query == "exit" {
			fmt.Println("Goodbye.")
			return
		}

		done := make(chan struct{})
		go displayProgress(done)

		res, err := sendReq("localhost", "30010", query)
		close(done)
		time.Sleep(50 * time.Millisecond)

		if err != nil {
			fmt.Printf("Error: %v", err)
		}
		fmt.Println(res)
	}
}

func displayProgress(done chan struct{}) {
	spinner := []string{"|", "/", "-", "\\"}
	i := 0

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			fmt.Print("\r\033[K")
			return
		case <-ticker.C:
			fmt.Printf("\rRetrieving output %s", spinner[i])
			i = (i + 1) % len(spinner)
		}
	}
}

// Send request to query pod
func sendReq(url string, port string, body string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s", url, port), bytes.NewBuffer([]byte(fmt.Sprintf("{\"body\": \"%s\"}", body))))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	} else if res.StatusCode != http.StatusOK {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return "", fmt.Errorf("Error reading response: %v", err)
		}
		return "", fmt.Errorf(string(resBody))
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	return string(resBody), nil
}
