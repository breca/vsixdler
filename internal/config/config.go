package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var validPlatforms = map[string]bool{
	"win32-x64":    true,
	"win32-arm64":  true,
	"linux-x64":    true,
	"linux-arm64":  true,
	"linux-armhf":  true,
	"alpine-x64":   true,
	"alpine-arm64": true,
	"darwin-x64":   true,
	"darwin-arm64": true,
	"web":          true,
}

type Extension struct {
	ID        string   `yaml:"id"`
	Version   string   `yaml:"version,omitempty"`
	Platforms []string `yaml:"platforms,omitempty"`
}

func (e Extension) Publisher() string {
	parts := strings.SplitN(e.ID, ".", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func (e Extension) Name() string {
	parts := strings.SplitN(e.ID, ".", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

type Config struct {
	Extensions []Extension `yaml:"extensions"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Extensions) == 0 {
		return fmt.Errorf("config: no extensions defined")
	}

	seen := make(map[string]bool)
	for i, ext := range c.Extensions {
		if ext.ID == "" {
			return fmt.Errorf("config: extension %d has no id", i)
		}
		if !strings.Contains(ext.ID, ".") {
			return fmt.Errorf("config: extension %q must be in publisher.name format", ext.ID)
		}
		if seen[ext.ID] {
			return fmt.Errorf("config: duplicate extension %q", ext.ID)
		}
		seen[ext.ID] = true

		for _, p := range ext.Platforms {
			if !validPlatforms[p] {
				return fmt.Errorf("config: extension %q has invalid platform %q", ext.ID, p)
			}
		}
	}

	return nil
}
