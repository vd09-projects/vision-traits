package traits

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vd09-projects/vision-traits/internal/config"
	"github.com/vd09-projects/vision-traits/internal/ollama"
	"github.com/vd09-projects/vision-traits/internal/prompt"
)

type ollamaClient interface {
	Generate(context.Context, ollama.GenerateRequest) (string, error)
}

type Extractor struct {
	cfg    config.Config
	ollama ollamaClient
}

func NewExtractor(cfg config.Config, client *ollama.Client) *Extractor {
	return &Extractor{cfg: cfg, ollama: client}
}

func (e *Extractor) Extract(ctx context.Context, imagesBase64 []string) (ExtractedTraits, error) {
	// Render prompt
	temp := prompt.TemplateData{
		Locale:   e.cfg.Traits.Locale,
		Taxonomy: e.cfg.Traits.Taxonomy,
		Vars:     e.cfg.Prompt.Vars,
	}
	p, err := temp.Render(e.cfg.Prompt.TemplatePath)
	if err != nil {
		return ExtractedTraits{}, err
	}

	// Ask Ollama to return schema-valid JSON.
	out, err := e.ollama.Generate(ctx, ollama.GenerateRequest{
		Prompt: p,
		Images: imagesBase64,
		Format: TraitsJSONSchema(),
	})
	if err != nil {
		return ExtractedTraits{}, err
	}

	var parsed ExtractedTraits
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return ExtractedTraits{}, fmt.Errorf("model returned non-json or schema-mismatch: %w; raw=%s", err, out)
	}

	// Optional: ensure all taxonomy keys exist (fill unknown).
	for _, t := range e.cfg.Traits.Taxonomy {
		if parsed.Traits == nil {
			parsed.Traits = map[string]TraitCategoryResult{}
		}
		if _, ok := parsed.Traits[t.Name]; !ok {
			parsed.Traits[t.Name] = TraitCategoryResult{
				Summary:    "unknown",
				Signals:    []string{},
				Confidence: 0,
			}
			parsed.Notes = append(parsed.Notes, "missing taxonomy key: "+t.Name)
		}
	}

	return parsed, nil
}
