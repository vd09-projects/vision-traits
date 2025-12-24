package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		temp    string
		data    any
		want    string
		wantErr string
	}{
		{
			name: "renders template with struct data",
			temp: "Locale: {{.Locale}}; Trait: {{index .Vars \"trait\"}}",
			data: TemplateData{
				Locale: "en-US",
				Vars: map[string]interface{}{
					"trait": "color",
				},
			},
			want: "Locale: en-US; Trait: color",
		},
		{
			name:    "errors when referenced key is missing",
			temp:    "Missing Msg: {{.Missing}}",
			data:    TemplateData{Locale: "en"},
			wantErr: "Missing",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := RenderTemplate(tt.temp, tt.data)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("RenderTemplate() unexpected error: %v", err)
				}
				if got != tt.want {
					t.Fatalf("RenderTemplate() = %q, want %q", got, tt.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("RenderTemplate() expected error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("RenderTemplate() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateDataRender(t *testing.T) {
	t.Parallel()

	t.Run("renders template file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "prompt.tmpl")
		const tmplBody = "Locale: {{.Locale}}, Trait: {{index .Vars \"trait\"}}"
		if err := os.WriteFile(path, []byte(tmplBody), 0o600); err != nil {
			t.Fatalf("failed to write template file: %v", err)
		}

		data := TemplateData{
			Locale: "hi-IN",
			Vars: map[string]interface{}{
				"trait": "shape",
			},
		}

		got, err := data.Render(path)
		if err != nil {
			t.Fatalf("TemplateData.Render() unexpected error: %v", err)
		}
		want := "Locale: hi-IN, Trait: shape"
		if got != want {
			t.Fatalf("TemplateData.Render() = %q, want %q", got, want)
		}
	})

	t.Run("returns error when file missing", func(t *testing.T) {
		data := TemplateData{}
		_, err := data.Render(filepath.Join(t.TempDir(), "missing.tmpl"))
		if err == nil {
			t.Fatalf("TemplateData.Render() expected error for missing file")
		}
	})

	t.Run("propagates template execution errors", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.tmpl")
		if err := os.WriteFile(path, []byte("{{.DoesNotExist}}"), 0o600); err != nil {
			t.Fatalf("failed to write template file: %v", err)
		}

		_, err := TemplateData{Locale: "fr-FR"}.Render(path)
		if err == nil {
			t.Fatalf("TemplateData.Render() expected template execution error")
		}
		if !strings.Contains(err.Error(), "DoesNotExist") {
			t.Fatalf("TemplateData.Render() error = %v, want substring %q", err, "DoesNotExist")
		}
	})
}
