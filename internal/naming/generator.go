package naming

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/google/uuid"

	"8stash/internal/config"
)

func BuildStashHash() (string, error){
	if config.NamingHashType == config.HashUUID {
		hash, err := buildUUIDHash()
		if err != nil {
			return "", err
		}
		return config.BranchPrefix + hash, nil
	}
	return config.BranchPrefix + buildNumericHash(), nil
}

func buildNumericHash() string {
	var randomInt = rand.Intn(config.HashRange)
	return strconv.Itoa(randomInt)
}

func buildUUIDHash() (string, error) {
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return randomUUID.String(), nil
}

