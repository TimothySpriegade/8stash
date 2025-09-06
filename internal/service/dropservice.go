package service

import "8stash/internal/gitx"

func HandleDrop(stashNr string) error {
	if err := gitx.DeleteBranch(BranchPrefix + stashNr); err != nil {
		return err
	}
	return nil
}
