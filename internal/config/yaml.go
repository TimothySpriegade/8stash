package config

import (
	"fmt"
	"os"
	"strings"

	yaml "go.yaml.in/yaml/v4"
)

type YamlConfig struct {
	CustomBranchPrefix string `yaml:"branch_prefix"`
	RetentionDays      int    `yaml:"retention_days"`
}

func (c *YamlConfig) sanitize() {
	c.CustomBranchPrefix = strings.TrimSpace(c.CustomBranchPrefix)
	c.CustomBranchPrefix = strings.Trim(c.CustomBranchPrefix, "/")
}

func (c *YamlConfig) validate() error {
	if c.RetentionDays < 0 {
		return fmt.Errorf("retention_days must be >= 0")
	}
	return nil
}

func LoadConfig(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("No 8stash config found")
		return nil
	}

	fmt.Println("trying to load 8stash config")
	var cfg YamlConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	cfg.sanitize()
	if err := cfg.validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	UpdateConstByConfig(&cfg)
	return nil
}
