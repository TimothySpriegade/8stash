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

func StashChangesToNewBranch(newBranchName string, commitMessage string) error {
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

	if err := stageChanges(wt); err != nil {
		return err
	}

	if err := commitChanges(repo, wt, newBranchName, commitMessage); err != nil {
		return err
	}

	if err := pushChanges(remote, repo, newBranchName); err != nil {
		return err
	}

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

func resolveAuthor(repo *git.Repository) (name, email string) {
	if cfg, err := repo.Config(); err == nil {
		name, email = cfg.User.Name, cfg.User.Email
	}
	if name == "" || email == "" {
		if globalCfg, err := config.LoadConfig(config.GlobalScope); err == nil {
			if name == "" {
				name = globalCfg.User.Name
			}
			if email == "" {
				email = globalCfg.User.Email
			}
		}
	}
	if name == "" {
		name = "8stash"
	}
	if email == "" {
		email = "noreply@local"
	}
	return
}

func commitChanges(repo *git.Repository, wt *git.Worktree, branchName string, commitMessage string) error {
	authorName, authorEmail := resolveAuthor(repo)

	if commitMessage == "" {
		commitMessage = fmt.Sprintf("move local changes to branch %s", branchName)
	}

	if _, err := wt.Commit(
		commitMessage,
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
