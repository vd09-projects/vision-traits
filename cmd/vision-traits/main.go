package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/vd09-projects/vision-traits/internal/config"
	"github.com/vd09-projects/vision-traits/internal/ollama"
	"github.com/vd09-projects/vision-traits/internal/traits"
	util "github.com/vd09-projects/vision-traits/internal/utils"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "configs/config.yaml", "path to yaml config")
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: vision-traits -config configs/config.yaml img1.jpg img2.jpg ...")
		os.Exit(2)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	paths, err = util.LimitSlice(paths, cfg.Traits.MaxImages)
	if err != nil {
		fmt.Fprintln(os.Stderr, "args error:", err)
		os.Exit(1)
	}

	var imgs []string
	for _, p := range paths {
		b64, err := util.ReadImageAsBase64(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "image error:", p, err)
			os.Exit(1)
		}
		imgs = append(imgs, b64)
	}

	client := ollama.New(cfg.Ollama.BaseURL, cfg.Ollama.Endpoint, cfg.Ollama.Model, cfg.Ollama.Timeout())
	extractor := traits.NewExtractor(cfg, client)

	ctx := context.Background()
	res, err := extractor.Extract(ctx, imgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "extract error:", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)
}
