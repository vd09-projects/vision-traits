package traits

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vd09-projects/vision-traits/config"
	itconfig "github.com/vd09-projects/vision-traits/internal/config"
	"github.com/vd09-projects/vision-traits/internal/ollama"
	ittraits "github.com/vd09-projects/vision-traits/internal/traits"
)

type stubOllama struct {
	t              *testing.T
	expectedImages []string
	response       string
}

func (s stubOllama) Generate(ctx context.Context, req ollama.GenerateRequest) (string, error) {
	if s.expectedImages != nil && !reflect.DeepEqual(req.Images, s.expectedImages) {
		s.t.Fatalf("Generate images = %v, want %v", req.Images, s.expectedImages)
	}
	return s.response, nil
}

func writeTemplate(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.tmpl")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write template: %v", err)
	}
	return path
}

func baseConfig(t *testing.T, maxImages int, tmplPath string) config.Config {
	t.Helper()
	return config.Config{
		Ollama: itconfig.OllamaConfig{
			BaseURL:   "http://example.com",
			Endpoint:  "/api",
			Model:     "model",
			TimeoutMS: 5000,
		},
		Traits: itconfig.TraitsConfig{
			MaxImages: maxImages,
			Locale:    "en-US",
			Taxonomy: []itconfig.TaxonomyItem{
				{Name: "color"},
				{Name: "shape"},
			},
		},
		Prompt: itconfig.PromptConfig{
			TemplatePath: tmplPath,
			Vars:         map[string]interface{}{"brand": "demo"},
		},
	}
}

func TestNewErrors(t *testing.T) {
	t.Parallel()

	if _, err := New(); err == nil {
		t.Fatalf("New() without config expected error")
	}

	if _, err := New(WithConfig(nil)); err == nil {
		t.Fatalf("New() with nil config expected error")
	}

	if _, err := New(WithConfigPath("missing-file")); err == nil {
		t.Fatalf("New() with missing config path expected error")
	}

	if _, err := New(WithOllamaClient(nil)); err == nil {
		t.Fatalf("New() with nil ollama client expected error")
	}
}

func TestExtractFromBase64LimitsAndDefaults(t *testing.T) {
	t.Parallel()

	tmpl := writeTemplate(t, "Locale: {{.Locale}}")
	cfg := baseConfig(t, 1, tmpl)

	resp := `{"global_confidence":90,"traits":{"color":{"summary":"red","signals":["bright"],"signals_by_key":{"hue":["red"]},"confidence":80}},"notes":["model-note"]}`
	client := stubOllama{t: t, expectedImages: []string{"img1"}, response: resp}

	vt, err := New(WithConfig(&cfg), WithOllamaClient(client))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	out, err := vt.ExtractFromBase64(context.Background(), []string{"img1", "img2"})
	if err != nil {
		t.Fatalf("ExtractFromBase64() unexpected error: %v", err)
	}

	res, ok := out.(ittraits.ExtractedTraits)
	if !ok {
		t.Fatalf("ExtractFromBase64() type = %T, want ExtractedTraits", out)
	}

	if res.GlobalConfidence != 90 {
		t.Fatalf("global confidence = %d, want 90", res.GlobalConfidence)
	}

	color := res.Traits["color"]
	if color.Summary != "red" || color.Confidence != 80 {
		t.Fatalf("color trait = %+v, want summary red confidence 80", color)
	}
	if !reflect.DeepEqual(color.SignalsByKey, map[string][]string{"hue": []string{"red"}}) {
		t.Fatalf("color signals_by_key = %+v", color.SignalsByKey)
	}

	shape := res.Traits["shape"]
	if shape.Summary != "unknown" || shape.Confidence != 0 || len(shape.Signals) != 0 || len(shape.SignalsByKey) != 0 {
		t.Fatalf("shape default = %+v, want unknown defaults", shape)
	}

	if len(res.Notes) != 2 || res.Notes[1] != "missing taxonomy key: shape" {
		t.Fatalf("notes = %v, want model-note and missing taxonomy note", res.Notes)
	}
}

func TestExtractFromPathsLimitsAndReads(t *testing.T) {
	t.Parallel()

	tmpl := writeTemplate(t, "plain")
	cfg := baseConfig(t, 1, tmpl)

	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.png")
	path2 := filepath.Join(dir, "b.png")
	if err := os.WriteFile(path1, []byte("one"), 0o600); err != nil {
		t.Fatalf("write file1: %v", err)
	}
	if err := os.WriteFile(path2, []byte("two"), 0o600); err != nil {
		t.Fatalf("write file2: %v", err)
	}
	encoded1 := base64.StdEncoding.EncodeToString([]byte("one"))

	resp := `{"global_confidence":50,"traits":{"color":{"summary":"blue","signals":["cool"],"signals_by_key":{"tone":["cool"]},"confidence":60}},"notes":[]}`
	client := stubOllama{t: t, expectedImages: []string{encoded1}, response: resp}

	vt, err := New(WithConfig(&cfg), WithOllamaClient(client))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	out, err := vt.ExtractFromPaths(context.Background(), []string{path1, path2})
	if err != nil {
		t.Fatalf("ExtractFromPaths() unexpected error: %v", err)
	}

	res := out.(ittraits.ExtractedTraits)
	if res.Traits["color"].Summary != "blue" {
		t.Fatalf("color summary = %q, want blue", res.Traits["color"].Summary)
	}
	if len(res.Traits["shape"].SignalsByKey) != 0 {
		t.Fatalf("shape signals_by_key = %+v, want empty map", res.Traits["shape"].SignalsByKey)
	}
}

func TestExtractFromBase64Errors(t *testing.T) {
	t.Parallel()

	var vt *VisionTraits
	if _, err := vt.ExtractFromBase64(context.Background(), []string{"img"}); err == nil {
		t.Fatalf("nil receiver expected error")
	}

	tmpl := writeTemplate(t, "plain")
	cfg := baseConfig(t, 1, tmpl)
	client := stubOllama{t: t, response: "{}"}
	vt, err := New(WithConfig(&cfg), WithOllamaClient(client))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if _, err := vt.ExtractFromBase64(nil, []string{"img"}); err == nil {
		t.Fatalf("nil context expected error")
	}
	if _, err := vt.ExtractFromBase64(context.Background(), nil); err == nil {
		t.Fatalf("empty images expected error")
	}
}
