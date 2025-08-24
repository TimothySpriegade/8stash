package service

import (
	"8stash/internal/gitx"
	"8stash/internal/naming"
)

func HandlePush() (string, error) {
	if err := gitx.PrepareRepository(); err != nil {
		return "", err
	}

	stashName := naming.BuildStashHash()

	err := gitx.StashChangesToNewBranch(stashName)
	if err != nil {
		return "", err
	}

	return stashName, nil
}
