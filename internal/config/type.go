package config

type Config struct {
	Ollama OllamaConfig `yaml:"ollama"`
	Traits TraitsConfig `yaml:"traits"`
	Prompt PromptConfig `yaml:"prompt"`
}
