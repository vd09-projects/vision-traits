package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL  string
	endpoint string
	model    string
	httpc    *http.Client
}

func New(baseURL, endpoint, model string, timeout time.Duration) *Client {
	return &Client{
		baseURL:  baseURL,
		endpoint: endpoint,
		model:    model,
		httpc: &http.Client{
			Timeout: timeout,
		},
	}
}

// GenerateRequest matches Ollama /api/generate with images + format.  [oai_citation:4‡Ollama](https://docs.ollama.com/api/generate?utm_source=chatgpt.com)
type GenerateRequest struct {
	Model   string      `json:"model"`
	Prompt  string      `json:"prompt"`
	Images  []string    `json:"images,omitempty"`
	Stream  bool        `json:"stream"`
	Format  interface{} `json:"format,omitempty"` // "json" OR JSON schema object
	Options interface{} `json:"options,omitempty"`
}

type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	req.Model = c.model
	req.Stream = false

	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.endpoint, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpc.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama http status=%d", resp.StatusCode)
	}

	var gr GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", err
	}
	return gr.Response, nil
}
