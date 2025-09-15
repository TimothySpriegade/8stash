package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateConstByConfig_NoChangesWhenEmpty(t *testing.T) {
	origPrefix := BranchPrefix
	origRetention := CleanUpTimeInDays
	t.Cleanup(func() {
		BranchPrefix = origPrefix
		CleanUpTimeInDays = origRetention
	})

	UpdateApplicationConfiguration(&YamlConfig{})

	assert.Equal(t, origPrefix, BranchPrefix)
	assert.Equal(t, origRetention, CleanUpTimeInDays)
}

func TestUpdateConstByConfig_CustomPrefixApplies(t *testing.T) {
	origPrefix := BranchPrefix
	origRetention := CleanUpTimeInDays
	t.Cleanup(func() {
		BranchPrefix = origPrefix
		CleanUpTimeInDays = origRetention
	})

	cfg := &YamlConfig{CustomBranchPrefix: "custom"}
	UpdateApplicationConfiguration(cfg)

	assert.Equal(t, "custom/", BranchPrefix)
	assert.Equal(t, origRetention, CleanUpTimeInDays)
}

func TestUpdateConstByConfig_TrimsSpacesFromPrefix(t *testing.T) {
	origPrefix := BranchPrefix
	origRetention := CleanUpTimeInDays
	t.Cleanup(func() {
		BranchPrefix = origPrefix
		CleanUpTimeInDays = origRetention
	})

	cfg := &YamlConfig{CustomBranchPrefix: "  spaced-prefix  "}
	UpdateApplicationConfiguration(cfg)

	assert.Equal(t, "spaced-prefix/", BranchPrefix)
}

func TestUpdateConstByConfig_RetentionAppliedWhenPositive(t *testing.T) {
	origPrefix := BranchPrefix
	origRetention := CleanUpTimeInDays
	t.Cleanup(func() {
		BranchPrefix = origPrefix
		CleanUpTimeInDays = origRetention
	})

	cfg := &YamlConfig{RetentionDays: 7}
	UpdateApplicationConfiguration(cfg)

	assert.Equal(t, 7, CleanUpTimeInDays)
	assert.Equal(t, origPrefix, BranchPrefix)
}

func TestUpdateConstByConfig_ZeroOrNegativeRetentionIgnored(t *testing.T) {
	origPrefix := BranchPrefix
	origRetention := CleanUpTimeInDays
	t.Cleanup(func() {
		BranchPrefix = origPrefix
		CleanUpTimeInDays = origRetention
	})

	// zero should be ignored
	UpdateApplicationConfiguration(&YamlConfig{RetentionDays: 0})
	assert.Equal(t, origRetention, CleanUpTimeInDays)

	// negative should also be ignored by this function (validation is elsewhere)
	UpdateApplicationConfiguration(&YamlConfig{RetentionDays: -10})
	assert.Equal(t, origRetention, CleanUpTimeInDays)
}

func TestUpdateApplicationConfiguration_AppliesNamingAndRetention(t *testing.T) {
    origPrefix := BranchPrefix
    origRetention := CleanUpTimeInDays
    origHashType := NamingHashType
    origHashRange := HashRange
    t.Cleanup(func() {
        BranchPrefix = origPrefix
        CleanUpTimeInDays = origRetention
        NamingHashType = origHashType
        HashRange = origHashRange
    })

    cfg := &YamlConfig{
        CustomBranchPrefix: "repo-prefix",
        RetentionDays:      7,
    }
    cfg.Naming.HashType = HashNumeric
    cfg.Naming.Range = 5000

    UpdateApplicationConfiguration(cfg)

    assert.Equal(t, "repo-prefix/", BranchPrefix)
    assert.Equal(t, 7, CleanUpTimeInDays)
    assert.Equal(t, HashNumeric, NamingHashType)
    assert.Equal(t, 5000, HashRange)
}

func TestUpdateApplicationConfiguration_UUIDIgnoresRangeAndSetsType(t *testing.T) {
    origHashType := NamingHashType
    origHashRange := HashRange
    t.Cleanup(func() {
        NamingHashType = origHashType
        HashRange = origHashRange
    })

    cfg := &YamlConfig{}
    cfg.Naming.HashType = HashUUID
    cfg.Naming.Range = 2

    UpdateApplicationConfiguration(cfg)

    assert.Equal(t, HashUUID, NamingHashType)
    assert.Equal(t, 9999, HashRange)
}

func TestUpdateApplicationConfiguration_IgnoresNonPositiveRange(t *testing.T) {
    origHashRange := HashRange
    origHashType := NamingHashType
    t.Cleanup(func() {
        HashRange = origHashRange
        NamingHashType = origHashType
    })

    cfg := &YamlConfig{}
    cfg.Naming.HashType = HashNumeric
    cfg.Naming.Range = 0 // non-positive should be ignored

    UpdateApplicationConfiguration(cfg)

    assert.Equal(t, HashNumeric, NamingHashType)
    assert.Equal(t, 9999, HashRange)
}