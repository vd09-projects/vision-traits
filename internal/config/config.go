package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type OllamaConfig struct {
	BaseURL   string `yaml:"base_url"`
	Model     string `yaml:"model"`
	Endpoint  string `yaml:"endpoint"`
	TimeoutMS int    `yaml:"timeout_ms"`
}

func (o OllamaConfig) Timeout() time.Duration {
	if o.TimeoutMS <= 0 {
		return 120 * time.Second
	}
	return time.Duration(o.TimeoutMS) * time.Millisecond
}

type TraitsConfig struct {
	MaxImages int            `yaml:"max_images"`
	Locale    string         `yaml:"locale"`
	Taxonomy  []TaxonomyItem `yaml:"taxonomy"`
}

type TaxonomyItem struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type PromptConfig struct {
	TemplatePath string                 `yaml:"template_path"`
	Vars         map[string]interface{} `yaml:"vars"`
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	// sensible defaults
	if cfg.Ollama.Endpoint == "" {
		cfg.Ollama.Endpoint = "/api/generate"
	}
	if cfg.Traits.MaxImages <= 0 {
		cfg.Traits.MaxImages = 3
	}
	if cfg.Traits.Locale == "" {
		cfg.Traits.Locale = "en-IN"
	}
	return cfg, nil
}
