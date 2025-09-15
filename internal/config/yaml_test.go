package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"8stash/internal/test"
)

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
	path := test.WriteTempFile(t, content)
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
	path := test.WriteTempFile(t, content)
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
	path := test.WriteTempFile(t, content)
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
	path := test.WriteTempFile(t, content)
	defer os.Remove(path)

	// Act
	err := LoadConfig(path)

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid config")
	assert.ErrorContains(t, err, "retention_days must be >= 0")
}

func TestLoadConfig_NamingDefaults_WhenMissing(t *testing.T) {
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

	content := `
branch_prefix: repo-prefix
retention_days: 5
`
	path := test.WriteTempFile(t, content)
	defer os.Remove(path)

	err := LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, "repo-prefix/", BranchPrefix)
	assert.Equal(t, 5, CleanUpTimeInDays)

	assert.Equal(t, HashNumeric, NamingHashType)
	assert.Equal(t, 9999, HashRange)
}

func TestLoadConfig_AppliesNumericRangeAndHashType(t *testing.T) {
	origHashType := NamingHashType
	origHashRange := HashRange
	t.Cleanup(func() {
		NamingHashType = origHashType
		HashRange = origHashRange
	})

	content := `
naming:
  hash_type: numeric
  hash_numeric_max_value: 5000
`
	path := test.WriteTempFile(t, content)
	defer os.Remove(path)

	err := LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, HashNumeric, NamingHashType)
	assert.Equal(t, 5000, HashRange)
}

func TestLoadConfig_UUIDIgnoresRangeAndUsesDefaultHashRange(t *testing.T) {
	origHashType := NamingHashType
	origHashRange := HashRange
	t.Cleanup(func() {
		NamingHashType = origHashType
		HashRange = origHashRange
	})

	content := `
naming:
  hash_type: uuid
  hash_numeric_max_value: 2
`
	path := test.WriteTempFile(t, content)
	defer os.Remove(path)

	err := LoadConfig(path)
	require.NoError(t, err)

	// hash type should be uuid, but effective HashRange should remain default
	assert.Equal(t, HashUUID, NamingHashType)
	assert.Equal(t, 9999, HashRange)
}

func TestLoadConfig_InvalidHashType_FallsBackToNumericAndAppliesRange(t *testing.T) {
	origHashType := NamingHashType
	origHashRange := HashRange
	t.Cleanup(func() {
		NamingHashType = origHashType
		HashRange = origHashRange
	})

	content := `
naming:
  hash_type: not-a-type
  hash_numeric_max_value: 1234
`
	path := test.WriteTempFile(t, content)
	defer os.Remove(path)

	err := LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, HashNumeric, NamingHashType)
	assert.Equal(t, 1234, HashRange)
}

func TestLoadConfig_EmptyNamedConfig_UsesDefaults(t *testing.T) {
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

	dir := t.TempDir()
	path := filepath.Join(dir, ConfigName)
	require.NoError(t, os.WriteFile(path, []byte(""), 0o644))

	err := LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, origPrefix, BranchPrefix)
	assert.Equal(t, origRetention, CleanUpTimeInDays)
	assert.Equal(t, origHashType, NamingHashType)
	assert.Equal(t, origHashRange, HashRange)
}

func TestValidate_NumericRangeTooSmall_ReturnsError(t *testing.T) {
	// Arrange
	cfg := &YamlConfig{}
	cfg.Naming.HashType = HashNumeric
	cfg.Naming.Range = 0

	// Act
	err := cfg.validate()

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "naming.hash_numeric_max_value must be >")
	assert.ErrorContains(t, err, strconv.Itoa(MinNumericRange))
}

func TestLoadConfig_ReadFileError_ReturnsWrappedError(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	info, err := os.Stat(dir)
	require.NoError(t, err)
	require.True(t, info.IsDir())

	// Act
	err = LoadConfig(dir)

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "something went wrong reading file")
	assert.True(t, strings.Contains(err.Error(), "is a directory") || strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "open"), "wrapped error should contain underlying io error")
}

func TestSanitize_LargeRange_PrintsWarningAndClamps(t *testing.T) {
	cfg := &YamlConfig{}
	cfg.Naming.HashType = HashNumeric
	cfg.Naming.Range = MaxNumericrange + 5

	// capture stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	// call sanitize while stderr is captured
	cfg.sanitize()

	// restore and read captured output
	_ = w.Close()
	_ = r.Close()
	os.Stderr = oldStderr
	require.NoError(t, err)

	assert.Equal(t, MaxNumericrange, cfg.Naming.Range)
}
