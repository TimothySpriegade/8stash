package helper

import (
	"errors"
	"fmt"
	"os"
	"stashpass/validation"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
)

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

func updateRepository() error {
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

func PrepareRepository() {
	if !validation.IsGitRepository() {
		fmt.Fprintln(os.Stderr, "Not a Git repository.")
		os.Exit(1)
	}

	if err := updateRepository(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if !validation.HasChanges() {
		fmt.Println("No changes to local repository.")
		os.Exit(0)
	}
}

func StashChangesToNewBranch(branchName string) error {
	repo, wt, origBranch, remote, err := getRepoContext()
	if err != nil {
		return err
	}
	if branchName == "" {
		return fmt.Errorf("branch name must not be empty")
	}
	if branchName == origBranch {
		return fmt.Errorf("target branch equals current branch")
	}

	// Ensure the target branch does not already exist.
	if _, err := repo.Reference(plumbing.NewBranchReferenceName(branchName), true); err == nil {
		return fmt.Errorf("branch %q already exists", branchName)
	} else if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return err
	}

	// Create and switch to the new branch, keeping current changes.
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
		Keep:   true,
	}); err != nil {
		return err
	}

	// Stage everything (adds, mods, deletions).
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

	// Commit on the new branch.
	if _, err := wt.Commit(
		fmt.Sprintf("move local changes to branch %s", branchName),
		&git.CommitOptions{
			Author: &object.Signature{
				Name:  "stashpass",
				Email: "noreply@local",
				When:  time.Now(),
			},
		},
	); err != nil {
		return err
	}

	// Push the new branch to its remote.
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

	// Switch back to the original branch, discarding working changes there.
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(origBranch),
		Force:  true,
		Keep:   false,
	}); err != nil {
		return err
	}

	return nil
}
