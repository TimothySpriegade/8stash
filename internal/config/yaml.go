package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	yaml "go.yaml.in/yaml/v4"
)

type HashType string


const (
	HashNumeric HashType = "numeric"
	HashUUID    HashType = "uuid"
)

type YamlConfig struct {
	CustomBranchPrefix string `yaml:"branch_prefix"`
	RetentionDays      int    `yaml:"retention_days"`
	Naming             struct {
		HashType HashType `yaml:"hash_type"`
		Range    int      `yaml:"hash_numeric_max_value"` // this is maxvalue so not a diget count
	} `yaml:"naming"`
}

func (c *YamlConfig) sanitize() {
	c.CustomBranchPrefix = strings.TrimSpace(c.CustomBranchPrefix)
    c.CustomBranchPrefix = strings.Trim(c.CustomBranchPrefix, "/")
    if c.Naming.HashType == "" {
        c.Naming.HashType = HashNumeric
    }

    if c.Naming.Range <= MinNumericRange {
        c.Naming.Range = HashRange
    }

    if c.Naming.HashType == HashNumeric && c.Naming.Range > MaxNumericrange {
        print("8stash config Numeric range is too big switching to default max range consider using hash_type UUID")
        c.Naming.Range = MaxNumericrange
    }
	
    if c.Naming.HashType == HashUUID {
        c.Naming.Range = HashRange
    }
}

func (c *YamlConfig) validate() error {
	if c.RetentionDays < 0 {
		return fmt.Errorf("retention_days must be >= 0")
	}

	if c.Naming.HashType != HashNumeric && c.Naming.HashType != HashUUID {
		print("hash_type has to be either numeric or uuid, setting config to default numeric")
		c.Naming.HashType = HashNumeric
	}

	if c.Naming.HashType == HashNumeric {
		if c.Naming.Range <= MinNumericRange{
			return fmt.Errorf("naming.hash_numeric_max_value must be > %s", strconv.Itoa(MinNumericRange))
		}
	}

	return nil
}

func LoadConfig(path string) error {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("something went wrong reading file %w", err)
	}

	var cfg YamlConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	cfg.sanitize()
	if err := cfg.validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	UpdateApplicationConfiguration(&cfg)
	return nil
}
