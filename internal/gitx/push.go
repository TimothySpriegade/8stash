package gitx

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
)

func StashChangesToNewBranch(newBranchName string) error {
	repo, wt, origBranch, remote, err := getRepoContext()
	if err != nil {
		return err
	}
	if err := validateBranch(newBranchName, origBranch, repo); err != nil {
		return err
	}

	if err := createNewBranchAndSwitch(newBranchName, wt); err != nil {
		return err
	}
	// Stage everything (adds, mods, deletions).
	if err := stageChanges(wt); err != nil {
		return err
	}
	// Commit on the new branch.
	if err := commitChanges(repo, wt, newBranchName); err != nil {
		return err
	}
	// Push the new branch to its remote.
	if err := pushChanges(remote, repo, newBranchName); err != nil {
		return err
	}
	// Switch back to the original branch, discarding working changes there.
	if err := switchToBranch(origBranch, wt); err != nil {
		return err
	}

	return nil
}

func validateBranch(branchName string, origBranch string, repo *git.Repository) error {
	if branchName == "" {
		return fmt.Errorf(branchNameMustNotEmptyErrorMsg)
	}
	if branchName == origBranch {
		return fmt.Errorf("target branch equals current branch")
	}

	if _, err := repo.Reference(plumbing.NewBranchReferenceName(branchName), true); err == nil {
		return fmt.Errorf("branch %q already exists", branchName)
	} else if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return err
	}

	return nil
}

func stageChanges(wt *git.Worktree) error {
	status, err := wt.Status()
	if err != nil {
		return err
	}
	for path, s := range status {
		// Stage deletions.
		if s.Worktree == git.Deleted || s.Staging == git.Deleted {
			if _, err := wt.Remove(path); err != nil {
				return err
			}
			continue
		}
		// Stage adds and modifications (includes untracked files).
		if _, err := wt.Add(path); err != nil {
			return err
		}
	}
	return nil
}

func commitChanges(repo *git.Repository,wt *git.Worktree, branchName string) error {
	var authorName string
    var authorEmail string
	cfg, err := repo.Config()

    if err != nil {
        fmt.Printf("Warning: failed to get git config, using default author: %v\n", err)
    } else {
        authorName = cfg.User.Name
        authorEmail = cfg.User.Email
    }

    if authorName == "" {
        authorName = "8stash"
    }

    if authorEmail == "" {
        authorEmail = "noreply@local"
    }
	
	if _, err := wt.Commit(
		fmt.Sprintf("move local changes to branch %s", branchName),
		&git.CommitOptions{
			Author: &object.Signature{
				Name:  authorName,
				Email: authorEmail,
				When:  time.Now(),
			},
		},
	); err != nil {
		return err
	}
	return nil
}

func pushChanges(remote string, repo *git.Repository, branchName string) error {
	pushOpts := &git.PushOptions{
		RemoteName: remote,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/heads/" + branchName + ":refs/heads/" + branchName)},
	}
	if auth, err := ssh.NewSSHAgentAuth("git"); err == nil && auth != nil {
		pushOpts.Auth = auth
	}
	if err := repo.Push(pushOpts); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("push failed: %w", err)
	}
	return nil
}

func createNewBranchAndSwitch(branchName string, wt *git.Worktree) error {
	return wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
		Keep:   true,
	})
}

func switchToBranch(branchName string, wt *git.Worktree) error {
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Force:  true,
		Keep:   false,
	}); err != nil {
		return err
	}
	return nil
}
