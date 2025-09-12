package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidOperation(t *testing.T) {
	validOps := []string{"push", "pop", "list", "drop", "help", "cleanup"}
	invalidOps := []string{"commit", "merge", "rebase", "status", "checkout", "invalidOp", ""}

	for _, op := range validOps {
		if !isValidOperation(op) {
			t.Errorf("Expected operation '%s' to be valid, but got invalid", op)
		}
	}

	for _, op := range invalidOps {
		if isValidOperation(op) {
			t.Errorf("Expected operation '%s' to be invalid, but got valid", op)
		}
	}
}

func TestStashNumberIsRequired(t *testing.T) {
	tests := []struct {
		operation string
		expected  bool
	}{
		{"push", false},
		{"pop", false},
		{"list", false},
		{"drop", true},
		{"help", false},
		{"cleanup", false},
	}

	for _, test := range tests {
		result := stashNumberIsRequiered(test.operation)
		if result != test.expected {
			t.Errorf("For operation '%s', expected %v but got %v", test.operation, test.expected, result)
		}
	}
}

func TestArgValidation_NoArgs_DefaultsToPush(t *testing.T) {
	// Arrange
	args := []string{}

	// Act
	op, num, err := ArgValidation(args)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "push", op)
	assert.Equal(t, 0, num)
}

func TestArgValidation_InvalidOperation_Error(t *testing.T) {
	// Arrange
	args := []string{"unknown"}

	// Act
	op, num, err := ArgValidation(args)

	// Assert
	require.Error(t, err)
	assert.Equal(t, "", op)
	assert.Equal(t, 0, num)
	assert.ErrorContains(t, err, "invalid operation")
}

func TestArgValidation_Drop_WithoutNumber_Error(t *testing.T) {
	// Arrange
	args := []string{"drop"}

	// Act
	op, num, err := ArgValidation(args)

	// Assert
	require.Error(t, err)
	assert.Equal(t, "", op)
	assert.Equal(t, 0, num)
	assert.ErrorContains(t, err, "operation requires a stash number")
}

func TestArgValidation_Drop_WithNumber_Succeeds(t *testing.T) {
	// Arrange
	args := []string{"drop", "5"}

	// Act
	op, num, err := ArgValidation(args)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "drop", op)
	assert.Equal(t, 5, num)
}

func TestArgValidation_Pop_WithNumber_Succeeds(t *testing.T) {
	// Arrange
	args := []string{"pop", "3"}

	// Act
	op, num, err := ArgValidation(args)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "pop", op)
	assert.Equal(t, 3, num)
}

func TestArgValidation_Pop_WithoutNumber_Succeeds(t *testing.T) {
	// Arrange
	args := []string{"pop"}

	// Act
	op, num, err := ArgValidation(args)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "pop", op)
	assert.Equal(t, 0, num)
}

func TestArgValidation_InvalidNumber_Error(t *testing.T) {
	// Arrange
	args := []string{"pop", "abc"}

	// Act
	op, num, err := ArgValidation(args)

	// Assert
	require.Error(t, err)
	assert.Equal(t, "", op)
	assert.Equal(t, 0, num)
	// strconv.Atoi error text includes "invalid syntax"
	assert.ErrorContains(t, err, "invalid syntax")
}

func TestArgValidation_CaseInsensitiveOperation(t *testing.T) {
	// Arrange
	args := []string{"HeLp"}

	// Act
	op, num, err := ArgValidation(args)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "help", op)
	assert.Equal(t, 0, num)
}

func TestArgValidation_Cleanup_NoNumber_Succeeds(t *testing.T) {
	// Arrange
	args := []string{"cleanup"}

	// Act
	op, num, err := ArgValidation(args)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "cleanup", op)
	assert.Equal(t, 0, num)
}
