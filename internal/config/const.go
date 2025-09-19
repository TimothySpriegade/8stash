package config

import (
	"math"
	"strings"
)

const ConfigName = ".8stash.yaml"
const MaxNumericrange = math.MaxInt32
const MinNumericRange = 1

var BranchPrefix = "8stash/"
var CleanUpTimeInDays = 30
var NamingHashType = HashNumeric
var HashRange = 9999
var SkipConfirmations = false

func UpdateApplicationConfiguration(cfg *YamlConfig) {
	updateBranchPrefix(cfg.CustomBranchPrefix)
	UpdateCleanupRetentionTime(cfg.RetentionDays)
	updateNamingHashType(cfg.Naming.HashType)
	updateHashRange(cfg.Naming.Range, cfg.Naming.HashType)
}

func updateHashRange(i int, ht HashType) {
	if i > MinNumericRange && ht == HashNumeric{
		HashRange = i
	}
}

func updateBranchPrefix(s string) {
	if p := strings.TrimSpace(s); p != "" {
		BranchPrefix = p + "/"
	}
}

func UpdateCleanupRetentionTime(i int) {
	if i > 0 {
		CleanUpTimeInDays = i
	}
}

func updateNamingHashType(h HashType) {
	if h != "" {
		NamingHashType = h
	}
}

func UpdateSkipConfirmations(y bool){
	SkipConfirmations = y
}