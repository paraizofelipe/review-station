package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	GitLab GitLabConfig `toml:"gitlab"`
	Repos  []Repo       `toml:"repo"`
}

type GitLabConfig struct {
	BaseURL string `toml:"base_url"`
	Token   string `toml:"token"`
}

type Repo struct {
	Name  string `toml:"name"`
	Path  string `toml:"path"`
	Local string `toml:"local"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "review-station", "config.toml")
}

func Load(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
