package naming

import (
	"math/rand"
	"strconv"

	"8stash/internal/config"
)

func BuildStashHash() string {
	var randomInt = rand.Intn(9999)
	var randomNumber = strconv.Itoa(randomInt)
	return config.BranchPrefix + randomNumber
}
