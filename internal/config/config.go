package config

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
)

var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

type Config struct {
	GitLab   GitLabConfig   `toml:"gitlab"`
	OpenCode OpenCodeConfig `toml:"opencode"`
	Repos    []Repo         `toml:"repo"`
}

type GitLabConfig struct {
	BaseURL string `toml:"base_url"`
	Token   string `toml:"token"`
}

type OpenCodeConfig struct {
	Command string `toml:"command"`
}

type Repo struct {
	Name            string `toml:"name"`
	Path            string `toml:"path"`
	Local           string `toml:"local"`
	OpenCodeCommand string `toml:"opencode_command"`
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
	cfg.GitLab.BaseURL = expandEnv(cfg.GitLab.BaseURL)
	cfg.GitLab.Token = expandEnv(cfg.GitLab.Token)
	cfg.OpenCode.Command = expandEnv(cfg.OpenCode.Command)
	for i := range cfg.Repos {
		cfg.Repos[i].Name = expandEnv(cfg.Repos[i].Name)
		cfg.Repos[i].Path = expandEnv(cfg.Repos[i].Path)
		cfg.Repos[i].Local = expandEnv(cfg.Repos[i].Local)
		cfg.Repos[i].OpenCodeCommand = expandEnv(cfg.Repos[i].OpenCodeCommand)
	}
	return cfg, nil
}

func expandEnv(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		return os.Getenv(match[2 : len(match)-1])
	})
}
