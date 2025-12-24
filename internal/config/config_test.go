package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOllamaConfigTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  OllamaConfig
		want time.Duration
	}{
		{
			name: "uses default when TimeoutMS not set",
			cfg:  OllamaConfig{},
			want: 120 * time.Second,
		},
		{
			name: "uses provided timeout",
			cfg:  OllamaConfig{TimeoutMS: 2500},
			want: 2500 * time.Millisecond,
		},
		{
			name: "non-positive timeout falls back to default",
			cfg:  OllamaConfig{TimeoutMS: -1},
			want: 120 * time.Second,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.cfg.Timeout(); got != tt.want {
				t.Fatalf("OllamaConfig.Timeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("loads config and applies defaults", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		const yamlBody = `
ollama:
  base_url: http://ollama.local
  model: llama3
  timeout_ms: 5000
traits:
  taxonomy:
    - name: size
      description: object size
prompt:
  template_path: "/tmp/template"
  vars:
    color: blue
`
		if err := os.WriteFile(path, []byte(yamlBody), 0o600); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() unexpected error: %v", err)
		}

		if cfg.Ollama.Endpoint != "/api/generate" {
			t.Fatalf("Load() default endpoint = %q, want %q", cfg.Ollama.Endpoint, "/api/generate")
		}
		if cfg.Traits.MaxImages != 3 {
			t.Fatalf("Load() default MaxImages = %d, want 3", cfg.Traits.MaxImages)
		}
		if cfg.Traits.Locale != "en-IN" {
			t.Fatalf("Load() default Locale = %q, want %q", cfg.Traits.Locale, "en-IN")
		}

		if cfg.Ollama.BaseURL != "http://ollama.local" {
			t.Fatalf("Load() base url = %q, want %q", cfg.Ollama.BaseURL, "http://ollama.local")
		}
		if cfg.Ollama.Model != "llama3" {
			t.Fatalf("Load() model = %q, want %q", cfg.Ollama.Model, "llama3")
		}
		if len(cfg.Traits.Taxonomy) != 1 {
			t.Fatalf("Load() taxonomy length = %d, want 1", len(cfg.Traits.Taxonomy))
		}
		if cfg.Traits.Taxonomy[0].Name != "size" {
			t.Fatalf("Load() taxonomy name = %q, want %q", cfg.Traits.Taxonomy[0].Name, "size")
		}
		if cfg.Prompt.TemplatePath != "/tmp/template" {
			t.Fatalf("Load() template path = %q, want %q", cfg.Prompt.TemplatePath, "/tmp/template")
		}
		if got := cfg.Prompt.Vars["color"]; got != "blue" {
			t.Fatalf("Load() vars[color] = %v, want %q", got, "blue")
		}
	})

	t.Run("returns error when file missing", func(t *testing.T) {
		t.Parallel()
		if _, err := Load(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
			t.Fatalf("Load() expected error for missing file")
		}
	})

	t.Run("returns error when yaml invalid", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.yaml")
		if err := os.WriteFile(path, []byte("traits: [invalid"), 0o600); err != nil {
			t.Fatalf("failed to write invalid yaml: %v", err)
		}
		if _, err := Load(path); err == nil {
			t.Fatalf("Load() expected yaml unmarshal error")
		}
	})
}
