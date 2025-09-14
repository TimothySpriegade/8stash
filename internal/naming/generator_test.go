package naming

import (
	"strconv"
	"testing"
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
