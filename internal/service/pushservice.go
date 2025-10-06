package service

import (
	"8stash/internal/gitx"
	"8stash/internal/naming"
)

func HandlePush(commitMessage string) (string, error) {
	if err := gitx.PrepareRepository(); err != nil {
		return "", err
	}

	stashName, err := naming.BuildStashHash()
	if err != nil {
		return "", err
	}

	err = gitx.StashChangesToNewBranch(stashName, commitMessage)
	if err != nil {
		return "", err
	}

	return stashName, nil
}
