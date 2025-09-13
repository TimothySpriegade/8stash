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

    UpdateConstByConfig(&YamlConfig{})

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
    UpdateConstByConfig(cfg)

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
    UpdateConstByConfig(cfg)

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
    UpdateConstByConfig(cfg)

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
    UpdateConstByConfig(&YamlConfig{RetentionDays: 0})
    assert.Equal(t, origRetention, CleanUpTimeInDays)

    // negative should also be ignored by this function (validation is elsewhere)
    UpdateConstByConfig(&YamlConfig{RetentionDays: -10})
    assert.Equal(t, origRetention, CleanUpTimeInDays)
}