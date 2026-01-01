package ollama

import (
	"context"
)

type OllamaClient interface {
	Generate(context.Context, GenerateRequest) (string, error)
}
