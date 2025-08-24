package naming

import (
	"math/rand"
	"strconv"
)

func BuildStashHash() string {
	var randomInt = rand.Intn(9999)
	var randomNumber = strconv.Itoa(randomInt)
	return "8stash/" + randomNumber
}
