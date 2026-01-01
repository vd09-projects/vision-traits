// visiontraits/vision_traits.go
package traits

import (
	"context"
	"fmt"

	"github.com/vd09-projects/vision-traits/config"
	"github.com/vd09-projects/vision-traits/internal/ollama"
	ittraits "github.com/vd09-projects/vision-traits/internal/traits"
	util "github.com/vd09-projects/vision-traits/internal/utils"
)

// VisionTraits is a reusable facade for extracting structured traits from images.
type VisionTraits struct {
	cfg       config.Config
	ollama    ollama.OllamaClient
	extractor *ittraits.Extractor
}

// Option configures VisionTraits.
type Option func(*options) error

type options struct {
	cfg *config.Config

	// Advanced override: provide a prebuilt Ollama client (e.g., custom transport / mocking).
	ollamaClient ollama.OllamaClient
}

// WithConfig uses an already loaded config (recommended if your app owns config lifecycle).
func WithConfig(cfg *config.Config) Option {
	return func(o *options) error {
		if cfg == nil {
			return fmt.Errorf("WithConfig: cfg is nil")
		}
		o.cfg = cfg
		return nil
	}
}

// WithConfigPath loads config from YAML and applies it to options.
func WithConfigPath(path string) Option {
	return func(o *options) error {
		if path == "" {
			return fmt.Errorf("WithConfigPath: empty path")
		}

		loaded, err := config.Load(path)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		// IMPORTANT: actually set it on o (your previous code did not).
		o.cfg = &loaded
		return nil
	}
}

// WithOllamaClient overrides the internal ollama client creation.
// Use only if you need custom transport settings or want to mock.
func WithOllamaClient(c ollama.OllamaClient) Option {
	return func(o *options) error {
		if c == nil {
			return fmt.Errorf("WithOllamaClient: client is nil")
		}
		o.ollamaClient = c
		return nil
	}
}

// New creates a VisionTraits service.
func New(opts ...Option) (*VisionTraits, error) {
	var o options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	if o.cfg == nil {
		return nil, fmt.Errorf("missing config: provide WithConfig(...) or WithConfigPath(...)")
	}
	cfg := *o.cfg

	var oc ollama.OllamaClient
	if o.ollamaClient != nil {
		oc = o.ollamaClient
	} else {
		oc = ollama.New(cfg.Ollama.BaseURL, cfg.Ollama.Endpoint, cfg.Ollama.Model, cfg.Ollama.Timeout())
	}

	extractor := ittraits.NewExtractor(cfg, oc)

	return &VisionTraits{
		cfg:       cfg,
		ollama:    oc,
		extractor: extractor,
	}, nil
}

// Config returns the underlying loaded config (read-only usage).
func (v *VisionTraits) Config() config.Config { return v.cfg }

// MaxImages returns configured max images (0 means no limit).
func (v *VisionTraits) MaxImages() int {
	if v == nil {
		return 0
	}
	return v.cfg.Traits.MaxImages
}

// ExtractFromBase64 extracts traits from base64-encoded images.
// It applies cfg.Traits.MaxImages limit (if > 0).
func (v *VisionTraits) ExtractFromBase64(ctx context.Context, base64Images []string) (any, error) {
	if v == nil || v.extractor == nil {
		return nil, fmt.Errorf("visiontraits not initialized")
	}
	if ctx == nil {
		return nil, fmt.Errorf("ctx is nil")
	}
	if len(base64Images) == 0 {
		return nil, fmt.Errorf("no images provided")
	}

	limited := base64Images
	if v.cfg.Traits.MaxImages > 0 {
		var err error
		limited, err = util.LimitSlice(base64Images, v.cfg.Traits.MaxImages)
		if err != nil {
			return nil, fmt.Errorf("limit images: %w", err)
		}
	}

	res, err := v.extractor.Extract(ctx, limited)
	if err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}
	return res, nil
}

// ExtractFromPaths reads images from disk (as base64) and extracts traits.
// It applies cfg.Traits.MaxImages limit (if > 0).
func (v *VisionTraits) ExtractFromPaths(ctx context.Context, paths []string) (any, error) {
	if v == nil || v.extractor == nil {
		return nil, fmt.Errorf("visiontraits not initialized")
	}
	if ctx == nil {
		return nil, fmt.Errorf("ctx is nil")
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no image paths provided")
	}

	limited := paths
	if v.cfg.Traits.MaxImages > 0 {
		var err error
		limited, err = util.LimitSlice(paths, v.cfg.Traits.MaxImages)
		if err != nil {
			return nil, fmt.Errorf("limit paths: %w", err)
		}
	}

	imgs := make([]string, 0, len(limited))
	for _, p := range limited {
		b64, err := util.ReadImageAsBase64(p)
		if err != nil {
			return nil, fmt.Errorf("read image %q: %w", p, err)
		}
		imgs = append(imgs, b64)
	}

	return v.ExtractFromBase64(ctx, imgs)
}
