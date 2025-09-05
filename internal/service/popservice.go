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
		if stashNumber == "0" {
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
	if stashNumber == "0" {
		if len(stashes) > 1 {
			return errors.New("multiple pops found an no stash number given")
		}
		for branchName := range stashes {
			if err := gitx.MergeStashIntoCurrentBranch(branchName); err != nil {
				return err
			}
			fmt.Println("popped stash: " + branchName)
			if err := gitx.DeleteBranch(branchName); err != nil {
				fmt.Println("failed to delete branch: " + branchName)
				return err
			}
			return nil
		}
		return nil
	}

	var branchName = BranchPrefix + stashNumber
	if err := gitx.MergeStashIntoCurrentBranch(branchName); err != nil {
		return err
	}

	fmt.Println("trying to delete branch: " + branchName)
	if err := gitx.DeleteBranch(branchName); err != nil {
		fmt.Println("failed to delete branch: " + branchName)
		return err
	}

	return nil
}
