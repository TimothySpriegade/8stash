package validation

import (
	"errors"
	"fmt"

	"os"

	"github.com/go-git/go-git/v6"
)

func IsGitRepository() bool {
	_, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if errors.Is(err, git.ErrRepositoryNotExists) {
		fmt.Println("Current directory is not a Git repository.")
		return false
	}
	fmt.Println("Current directory is a valid Git repository.")
	return true
}

func HasChanges() bool {
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository: %v\n", err)
		return false
	}

	worktree, err := repo.Worktree()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting worktree: %v\n", err)
		return false
	}

	status, err := worktree.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		return false
	}

	return !status.IsClean()
}
