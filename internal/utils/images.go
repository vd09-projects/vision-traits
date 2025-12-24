package util

import (
	"encoding/base64"
	"fmt"
	"os"
)

func ReadImageAsBase64(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Ollama REST API expects base64-encoded image bytes in the `images` array.  [oai_citation:2‡Ollama](https://docs.ollama.com/capabilities/vision?utm_source=chatgpt.com)
	return base64.StdEncoding.EncodeToString(b), nil
}

func LimitSlice[T any](items []T, n int) ([]T, error) {
	if n <= 0 {
		return nil, fmt.Errorf("limit must be > 0")
	}
	if len(items) <= n {
		return items, nil
	}
	return items[:n], nil
}
