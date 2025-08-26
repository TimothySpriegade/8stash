package validation

import (
	"testing"
)

func TestIsvalidOperation(t *testing.T) {
	validOps := []string{"push", "pop", "list", "drop", "help"}
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
	}

	for _, test := range tests {
		result := stashNumberIsRequiered(test.operation)
		if result != test.expected {
			t.Errorf("For operation '%s', expected %v but got %v", test.operation, test.expected, result)
		}
	}
}
