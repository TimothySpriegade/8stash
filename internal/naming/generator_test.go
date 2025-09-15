package naming

import (
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"8stash/internal/config"
)

func TestBuildStashHash(t *testing.T) {
	stashHash, err := BuildStashHash()
	if err != nil {
		t.Error("Error building Hash")
	}

	if len(stashHash) < 7 || len(stashHash) > 13 {
		t.Errorf("Expected stash hash length between 7 and 13, got %d", len(stashHash))
	}

	if stashHash[:7] != "8stash/" {
		t.Errorf("Expected stash hash to start with '8stash/', got %s", stashHash)
	}

	numberPart := stashHash[7:]
	_, err = strconv.Atoi(numberPart) // use = instead of := to avoid redeclaring err
	if err != nil {
		t.Errorf("Expected number part to be valid integer, got '%s' with error: %v", numberPart, err)
	}
}

func TestBuildStashHash_UUIDFormat(t *testing.T) {
	origType := config.NamingHashType
	origPrefix := config.BranchPrefix
	t.Cleanup(func() {
		config.NamingHashType = origType
		config.BranchPrefix = origPrefix
	})

	config.NamingHashType = config.HashUUID
	config.BranchPrefix = "8stash/"

	h, err := BuildStashHash()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(h, config.BranchPrefix))
	u := strings.TrimPrefix(h, config.BranchPrefix)
	_, err = uuid.Parse(u)
	require.NoError(t, err, "expected valid UUID suffix")
}

func TestBuildStashHash_NumericRangeAndDigits(t *testing.T) {
	origType := config.NamingHashType
	origRange := config.HashRange
	origPrefix := config.BranchPrefix
	t.Cleanup(func() {
		config.NamingHashType = origType
		config.HashRange = origRange
		config.BranchPrefix = origPrefix
	})

	config.NamingHashType = config.HashNumeric
	config.HashRange = 1000
	config.BranchPrefix = "8stash/"

	// sample multiple times to exercise randomness
	for i := 0; i < 50; i++ {
		h, err := BuildStashHash()
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(h, config.BranchPrefix))
		numPart := strings.TrimPrefix(h, config.BranchPrefix)
		n, err := strconv.Atoi(numPart)
		require.NoError(t, err, "numeric suffix should parse")
		assert.GreaterOrEqual(t, n, 0)
		assert.Less(t, n, config.HashRange)
	}
}

func TestBuildStashHash_RangeOneProducesZero(t *testing.T) {
	origType := config.NamingHashType
	origRange := config.HashRange
	t.Cleanup(func() {
		config.NamingHashType = origType
		config.HashRange = origRange
	})

	config.NamingHashType = config.HashNumeric
	config.HashRange = 1
	config.BranchPrefix = "8stash/"

	h, err := BuildStashHash()
	require.NoError(t, err)
	numPart := strings.TrimPrefix(h, config.BranchPrefix)
	assert.Equal(t, "0", numPart, "with range 1 rand.Intn(1) yields 0")
}
