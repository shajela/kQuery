package embeddings

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"
)

// Send request for embeddings
func ReqEmb(url string, payload []byte, headers map[string]string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if !slices.Contains([]int{http.StatusOK, http.StatusAccepted, http.StatusCreated}, res.StatusCode) {
		return nil, fmt.Errorf("HTTP error: recieved status code %d\n%s", res.StatusCode, string(body))
	}

	return body, nil
}
