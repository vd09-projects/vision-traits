package traits

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/vd09-projects/vision-traits/internal/config"
	"github.com/vd09-projects/vision-traits/internal/ollama"
)

type stubGenerator func(context.Context, ollama.GenerateRequest) (string, error)

func (s stubGenerator) Generate(ctx context.Context, req ollama.GenerateRequest) (string, error) {
	return s(ctx, req)
}

func TestExtractorExtractSuccess(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	templatePath := filepath.Join(dir, "prompt.tmpl")
	const tmpl = `Locale: {{.Locale}} {{index .Vars "greet"}}`
	if err := os.WriteFile(templatePath, []byte(tmpl), 0o600); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	cfg := config.Config{
		Traits: config.TraitsConfig{
			Locale: "en-US",
			Taxonomy: []config.TaxonomyItem{
				{Name: "color"},
				{Name: "shape"},
			},
		},
		Prompt: config.PromptConfig{
			TemplatePath: templatePath,
			Vars: map[string]interface{}{
				"greet": "hello",
			},
		},
	}

	images := []string{"img-base64"}
	expectedPrompt := "Locale: en-US hello"

	stub := stubGenerator(func(ctx context.Context, req ollama.GenerateRequest) (string, error) {
		if req.Prompt != expectedPrompt {
			t.Fatalf("Generate prompt = %q, want %q", req.Prompt, expectedPrompt)
		}
		if !reflect.DeepEqual(req.Images, images) {
			t.Fatalf("Generate images = %v, want %v", req.Images, images)
		}
		format, ok := req.Format.(map[string]interface{})
		if !ok || format["type"] != "object" {
			t.Fatalf("Generate format = %#v, want schema map", req.Format)
		}
		return `{"global_confidence":80,"traits":{"color":{"summary":"red","signals":["warm"],"signals_by_key":{"tone":["warm"]},"confidence":70}},"notes":["from-model"]}`, nil
	})

	ext := &Extractor{cfg: cfg, ollama: stub}

	got, err := ext.Extract(context.Background(), images)
	if err != nil {
		t.Fatalf("Extract() unexpected error: %v", err)
	}

	if got.GlobalConfidence != 80 {
		t.Fatalf("Extract() global confidence = %d, want 80", got.GlobalConfidence)
	}
	if got.Traits["color"].Summary != "red" {
		t.Fatalf("Extract() color summary = %q, want %q", got.Traits["color"].Summary, "red")
	}
	wantSignalsByKey := map[string][]string{"tone": []string{"warm"}}
	if !reflect.DeepEqual(got.Traits["color"].SignalsByKey, wantSignalsByKey) {
		t.Fatalf("Extract() color signals_by_key = %+v, want %+v", got.Traits["color"].SignalsByKey, wantSignalsByKey)
	}
	shape, ok := got.Traits["shape"]
	if !ok {
		t.Fatalf("Extract() expected missing taxonomy to be added")
	}
	if shape.Summary != "unknown" || shape.Confidence != 0 || len(shape.Signals) != 0 || len(shape.SignalsByKey) != 0 {
		t.Fatalf("Extract() defaulted trait = %+v, want unknown defaults", shape)
	}
	if len(got.Notes) != 2 {
		t.Fatalf("Extract() notes = %v, want two entries", got.Notes)
	}
	if got.Notes[1] != "missing taxonomy key: shape" {
		t.Fatalf("Extract() missing taxonomy note = %q, want %q", got.Notes[1], "missing taxonomy key: shape")
	}
}

func TestExtractorExtractErrors(t *testing.T) {
	t.Parallel()

	t.Run("template render error", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			Traits: config.TraitsConfig{Locale: "en-US"},
			Prompt: config.PromptConfig{
				TemplatePath: filepath.Join(t.TempDir(), "missing.tmpl"),
			},
		}

		stub := stubGenerator(func(ctx context.Context, req ollama.GenerateRequest) (string, error) {
			t.Fatalf("Generate should not be called on template error")
			return "", nil
		})

		ext := &Extractor{cfg: cfg, ollama: stub}
		if _, err := ext.Extract(context.Background(), nil); err == nil {
			t.Fatalf("Extract() expected error when template missing")
		}
	})

	t.Run("ollama error propagated", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		templatePath := filepath.Join(dir, "prompt.tmpl")
		if err := os.WriteFile(templatePath, []byte("plain"), 0o600); err != nil {
			t.Fatalf("failed to write template: %v", err)
		}
		cfg := config.Config{
			Traits: config.TraitsConfig{Locale: "en-US"},
			Prompt: config.PromptConfig{
				TemplatePath: templatePath,
			},
		}

		wantErr := errors.New("ollama failure")
		stub := stubGenerator(func(ctx context.Context, req ollama.GenerateRequest) (string, error) {
			return "", wantErr
		})

		ext := &Extractor{cfg: cfg, ollama: stub}
		_, err := ext.Extract(context.Background(), nil)
		if !errors.Is(err, wantErr) {
			t.Fatalf("Extract() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("invalid json from model", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		templatePath := filepath.Join(dir, "prompt.tmpl")
		if err := os.WriteFile(templatePath, []byte("plain"), 0o600); err != nil {
			t.Fatalf("failed to write template: %v", err)
		}
		cfg := config.Config{
			Traits: config.TraitsConfig{Locale: "en-US"},
			Prompt: config.PromptConfig{TemplatePath: templatePath},
		}

		stub := stubGenerator(func(ctx context.Context, req ollama.GenerateRequest) (string, error) {
			return "not-json", nil
		})

		ext := &Extractor{cfg: cfg, ollama: stub}
		_, err := ext.Extract(context.Background(), nil)
		if err == nil || !strings.Contains(err.Error(), "model returned non-json") {
			t.Fatalf("Extract() error = %v, want non-json message", err)
		}
	})
}

func TestTraitsJSONSchema(t *testing.T) {
	t.Parallel()

	schema := TraitsJSONSchema()

	if schema["type"] != "object" {
		t.Fatalf("schema type = %v, want object", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("schema properties missing or wrong type: %#v", schema["properties"])
	}

	for _, key := range []string{"global_confidence", "traits", "notes"} {
		if _, ok := props[key]; !ok {
			t.Fatalf("schema missing property %q", key)
		}
	}

	traitsProp, ok := props["traits"].(map[string]interface{})
	if !ok {
		t.Fatalf("schema traits property wrong type: %#v", props["traits"])
	}
	traitItem, ok := traitsProp["additionalProperties"].(map[string]interface{})
	if !ok {
		t.Fatalf("schema traits additionalProperties wrong type: %#v", traitsProp["additionalProperties"])
	}
	traitProps, ok := traitItem["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("schema trait properties wrong type: %#v", traitItem["properties"])
	}
	if _, ok := traitProps["signals_by_key"]; !ok {
		t.Fatalf("schema trait missing property %q", "signals_by_key")
	}

	required, ok := schema["required"].([]string)
	if !ok || len(required) == 0 {
		t.Fatalf("schema required invalid: %#v", schema["required"])
	}
}
