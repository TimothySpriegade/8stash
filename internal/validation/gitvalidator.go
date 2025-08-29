package validation

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v6"
)

const OpeningRepoErrorMessage = "Error opening repository:"

func IsGitRepository() error {
	_, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return err
	}
	fmt.Println("Current directory is a valid Git repository.")
	return nil
}

func HasChanges() error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf(OpeningRepoErrorMessage+" %v\n", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf(OpeningRepoErrorMessage+" %v\n", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf(OpeningRepoErrorMessage+" %v\n", err)
	}

	if status.IsClean() {
		return errors.New("no changes detected in working tree")
	}

	return nil
}
