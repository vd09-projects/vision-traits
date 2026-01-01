// package config
package config

import "github.com/vd09-projects/vision-traits/internal/config"

type Config = config.Config

func Load(path string) (Config, error) {
	return config.Load(path)
}
