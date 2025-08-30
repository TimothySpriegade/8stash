package service

import (
	"8stash/internal/gitx"
	"8stash/internal/validation"
	"errors"
	"fmt"
)

func HandlePop(stashNumber string) error {
	if err := validation.IsGitRepository(); err != nil {
		return err
	}

	if err := gitx.UpdateRepository(); err != nil {
		return err
	}

	stashes, err := Retrieve8stashList()
	if err != nil {
		return err
	}

	if len(stashes) == 0 {
		return errors.New("no pops found")
	}

	if len(stashes) > 1 {
		if stashNumber == "" {
			return errors.New("multiple pops found an no stash number given")
		}

		if err := popStash(stashNumber, stashes); err != nil {
			return err
		}
		return nil
	}

	if err := popStash(stashNumber, stashes); err != nil {
		return err
	}
	return nil
}

func popStash(stashNumber string, stashes map[string]string) error {
	if stashNumber == "" {
		for stashNr, _ := range stashes {
			var branchName = BranchPrefix + stashNr
			if err := gitx.MergeStashIntoCurrentBranch(branchName); err != nil {
				return err
			}
		}
	}

	var branchName = BranchPrefix + stashNumber
	if err := gitx.MergeStashIntoCurrentBranch(branchName); err != nil {
		return err
	}

	fmt.Println("trying to delete branch: " + branchName)
	if err := gitx.DeleteBranch(branchName); err != nil {
		return err
	}

	return nil
}
