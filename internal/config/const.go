package config

import (
	"strings"
)

const ConfigName = ".8stash.yaml"

var BranchPrefix = "8stash/"
var CleanUpTimeInDays = 30

func UpdateConstByConfig(cfg *YamlConfig) {
	if p := strings.TrimSpace(cfg.CustomBranchPrefix); p != "" {
		BranchPrefix = p + "/"
	}

	if cfg.RetentionDays > 0 {
		CleanUpTimeInDays = cfg.RetentionDays
	}
}
