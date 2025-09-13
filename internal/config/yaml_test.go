package config

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func writeTempFile(t *testing.T, content string) string {
    t.Helper()
    f, err := os.CreateTemp("", "8stash-config-*.yaml")
    require.NoError(t, err)
    require.NoError(t, os.WriteFile(f.Name(), []byte(content), 0o644))
    require.NoError(t, f.Close())
    return f.Name()
}

func TestLoadConfig_MissingFile_NoError_DefaultsUnchanged(t *testing.T) {
    // Arrange
    origPrefix := BranchPrefix
    origRetention := CleanUpTimeInDays
    t.Cleanup(func() {
        BranchPrefix = origPrefix
        CleanUpTimeInDays = origRetention
    })

    nonExistent := filepath.Join(os.TempDir(), "definitely-does-not-exist-8stash.yaml")

    // Act
    err := LoadConfig(nonExistent)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, origPrefix, BranchPrefix)
    assert.Equal(t, origRetention, CleanUpTimeInDays)
}

func TestLoadConfig_ValidFile_AppliesValues(t *testing.T) {
    // Arrange
    origPrefix := BranchPrefix
    origRetention := CleanUpTimeInDays
    t.Cleanup(func() {
        BranchPrefix = origPrefix
        CleanUpTimeInDays = origRetention
    })

    content := `
branch_prefix: customprefix
retention_days: 7
`
    path := writeTempFile(t, content)
    defer os.Remove(path)

    // Act
    err := LoadConfig(path)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "customprefix/", BranchPrefix)
    assert.Equal(t, 7, CleanUpTimeInDays)
}

func TestLoadConfig_SanitizesBranchPrefix_RemovesSlashesAndSpaces(t *testing.T) {
    // Arrange
    origPrefix := BranchPrefix
    t.Cleanup(func() { BranchPrefix = origPrefix })

    content := `
branch_prefix: " /my-prefix/ "
`
    path := writeTempFile(t, content)
    defer os.Remove(path)

    // Act
    err := LoadConfig(path)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "my-prefix/", BranchPrefix)
}

func TestLoadConfig_InvalidYAML_ReturnsError(t *testing.T) {
    // Arrange
    content := `: ::: not yaml`
    path := writeTempFile(t, content)
    defer os.Remove(path)

    // Act
    err := LoadConfig(path)

    // Assert
    require.Error(t, err)
    assert.ErrorContains(t, err, "parsing config")
}

func TestLoadConfig_NegativeRetention_ReturnsError(t *testing.T) {
    // Arrange
    origRetention := CleanUpTimeInDays
    t.Cleanup(func() { CleanUpTimeInDays = origRetention })

    content := `
retention_days: -5
`
    path := writeTempFile(t, content)
    defer os.Remove(path)

    // Act
    err := LoadConfig(path)

    // Assert
    require.Error(t, err)
    assert.ErrorContains(t, err, "invalid config")
    assert.ErrorContains(t, err, "retention_days must be >= 0")
}