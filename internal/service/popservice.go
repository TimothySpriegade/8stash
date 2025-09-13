package service

import (
	"errors"
	"fmt"

	"8stash/internal/gitx"
	"8stash/internal/validation"
	"8stash/internal/config"
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
			return errors.New("multiple stashes found; please specify which one to pop")
		}
		for branchName := range stashes {
			return applyAndRemoveStash(branchName)
		}
		fmt.Println("No stashes to pop.")
		return nil
	}

	branchName := config.BranchPrefix + stashNumber
	return applyAndRemoveStash(branchName)
}

func applyAndRemoveStash(branchName string) error {
	err := gitx.MergeStashIntoCurrentBranch(branchName)
	if err != nil {
		if errors.Is(err, gitx.ErrNonFastForward) {
			fmt.Println("Branches have diverged, attempting a three-way merge...")
			if mergeErr := gitx.ApplyDivergedMerge(branchName); mergeErr != nil {
				return mergeErr
			}
		} else {
			return err
		}
	}
	fmt.Println("Popped stash from branch: " + branchName)
	if err := gitx.DeleteBranch(branchName); err != nil {
		fmt.Println("Warning: failed to delete stash branch: " + branchName)
		return err
	}
	return nil
}
