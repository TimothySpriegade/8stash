package service

import (
	"math/rand"
	"stashpass/helper"
	"strconv"
)

func HandlePush() string {
	helper.PrepareRepository()
	stashName := buildStashHash()
	err := helper.StashChangesToNewBranch(stashName)
	if err != nil {
		panic(err)
	}

	return stashName
}

func buildStashHash() string {
	var randomInt = rand.Intn(9999)
	var randomNumber = strconv.Itoa(randomInt)
	return "stashpass/" + randomNumber
}
