package util

import (
	"errors"
	"fmt"
	"os"
	"stashpass/validation"

	"github.com/go-git/go-git/v6"
)

func updateRepository() {
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository: %v\n", err)
		os.Exit(1)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting worktree: %v\n", err)
		os.Exit(1)
	}

	err = worktree.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		fmt.Fprintf(os.Stderr, "Error pulling changes: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Repository updated successfully.")
}

func PrepareRepository() {
	if !validation.IsGitRepository() {
		os.Exit(1)
	}

	updateRepository()

	if !validation.HasChanges() {
		fmt.Println("No changes to local repostory.")
		os.Exit(0)
	}
}
