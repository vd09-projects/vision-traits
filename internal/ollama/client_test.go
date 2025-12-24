package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientGenerateSuccess(t *testing.T) {
	t.Parallel()

	const endpoint = "/api/generate"
	c := New("http://example.com", endpoint, "vision", time.Second)
	c.httpc = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			if r.URL.String() != "http://example.com/api/generate" {
				t.Fatalf("url = %s, want http://example.com/api/generate", r.URL.String())
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/json" {
				t.Fatalf("content-type = %s, want application/json", ct)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}
			var req GenerateRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("failed to unmarshal body: %v", err)
			}
			if req.Model != "vision" {
				t.Fatalf("model = %q, want %q", req.Model, "vision")
			}
			if req.Stream {
				t.Fatalf("stream flag should be false")
			}
			if req.Prompt != "describe image" {
				t.Fatalf("prompt = %q, want %q", req.Prompt, "describe image")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"response":"ok","done":true}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	resp, err := c.Generate(context.Background(), GenerateRequest{
		Prompt: "describe image",
		Stream: true,
	})
	if err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("Generate() = %q, want %q", resp, "ok")
	}
}

func TestClientGenerateHTTPErrors(t *testing.T) {
	t.Parallel()

	const endpoint = "/api/generate"
	tests := []struct {
		name    string
		rt      roundTripFunc
		wantErr string
	}{
		{
			name: "non-2xx status",
			rt: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadGateway,
					Body:       io.NopCloser(strings.NewReader("bad gateway")),
					Header:     make(http.Header),
				}, nil
			},
			wantErr: "status=502",
		},
		{
			name: "invalid json response",
			rt: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"response":`)),
					Header:     make(http.Header),
				}, nil
			},
			wantErr: "unexpected EOF",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := New("http://example.com", endpoint, "vision", time.Second)
			c.httpc = &http.Client{Transport: tt.rt}

			_, err := c.Generate(context.Background(), GenerateRequest{Prompt: "anything"})
			if err == nil {
				t.Fatalf("Generate() expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Generate() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestClientGenerateRequestErrors(t *testing.T) {
	t.Parallel()

	c := &Client{
		baseURL:  "://bad url",
		endpoint: "/generate",
		model:    "vision",
		httpc:    &http.Client{},
	}

	_, err := c.Generate(context.Background(), GenerateRequest{})
	if err == nil {
		t.Fatalf("Generate() expected error for invalid request URL")
	}
	if !strings.Contains(err.Error(), "missing protocol scheme") {
		t.Fatalf("Generate() error = %v, want missing scheme", err)
	}
}

func TestClientGenerateMarshalError(t *testing.T) {
	t.Parallel()

	c := &Client{
		baseURL:  "http://example.com",
		endpoint: "/generate",
		model:    "vision",
		httpc:    &http.Client{},
	}

	_, err := c.Generate(context.Background(), GenerateRequest{
		Format: func() {},
	})
	if err == nil {
		t.Fatalf("Generate() expected json marshal error")
	}
	var jsonErr *json.UnsupportedTypeError
	if !errors.As(err, &jsonErr) {
		t.Fatalf("Generate() error = %v, want UnsupportedTypeError", err)
	}
}
