package gitx

import (
	"8stash/internal/validation"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

const branchNameMustNotEmptyErrorMsg = "branch name must not be empty"

func getRepoContext() (*git.Repository, *git.Worktree, string, string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("open repo: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("worktree: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("HEAD: %w", err)
	}
	if !head.Name().IsBranch() {
		return nil, nil, "", "", fmt.Errorf("detached HEAD: cannot operate on current branch")
	}
	branch := head.Name().Short()

	remote := "origin"
	if cfg, _ := repo.Config(); cfg != nil {
		if b, ok := cfg.Branches[branch]; ok && b.Remote != "" {
			remote = b.Remote
		}
	}

	return repo, wt, branch, remote, nil
}

func PrepareRepository() error {
	if err := validation.IsGitRepository(); err != nil {
		return err
	}

	if err := UpdateRepository(); err != nil {
		return err
	}

	if err := validation.HasChanges(); err != nil {
		return err
	}

	return nil
}

func UpdateRepository() error {
	_, wt, branch, remote, err := getRepoContext()
	if err != nil {
		return err
	}

	err = wt.Pull(&git.PullOptions{
		RemoteName:    remote,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	switch {
	case err == nil:
		return nil
	case errors.Is(err, git.NoErrAlreadyUpToDate):
		return nil
	case errors.Is(err, git.ErrNonFastForwardUpdate):
		return fmt.Errorf("non fast-forward: local branch diverged from %s/%s", remote, branch)
	default:
		return fmt.Errorf("pull failed: %w", err)
	}
}
