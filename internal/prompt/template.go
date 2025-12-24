package prompt

import (
	"bytes"
	"os"
	"text/template"

	"github.com/vd09-projects/vision-traits/internal/config"
)

type TemplateData struct {
	Locale   string
	Taxonomy []config.TaxonomyItem
	Vars     map[string]interface{}
}

func (d TemplateData) Render(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return RenderTemplate(string(raw), d)
}

func RenderTemplate(temp string, data any) (string, error) {
	tmpl, err := template.New("traits").Option("missingkey=error").Parse(temp)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
