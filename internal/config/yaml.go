package config

import (
	"fmt"
	"os"
	"strings"

	yaml "go.yaml.in/yaml/v4"
)

type HashType string

const (
	HashNumeric HashType = "numeric"
	HashUUID HashType = "uuid"
)

type YamlConfig struct {
    CustomBranchPrefix string `yaml:"branch_prefix"`
    RetentionDays      int    `yaml:"retention_days"`
    Naming struct {
        HashType HashType `yaml:"hash_type"`
        Length   int      `yaml:"length"`
    } `yaml:"naming"`
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
		return nil
	}

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
