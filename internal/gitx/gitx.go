package gitx

import (
	"8stash/internal/validation"
	"errors"
	"fmt"
	"strings"
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

func validateBranch(branchName string, origBranch string, repo *git.Repository) error {
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

func commitChanges(wt *git.Worktree, branchName string) error {
	if _, err := wt.Commit(
		fmt.Sprintf("move local changes to branch %s", branchName),
		&git.CommitOptions{
			Author: &object.Signature{
				Name:  "8stash",
				Email: "noreply@local",
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

func StashChangesToNewBranch(newBranchName string) error {
	repo, wt, origBranch, remote, err := getRepoContext()
	if err != nil {
		return err
	}
	if err := validateBranch(newBranchName, origBranch, repo); err != nil {
		return err
	}
	// Create and switch to the new branch, keeping current changes.
	if err := switchToBranch(newBranchName, wt); err != nil {
		return err
	}
	// Stage everything (adds, mods, deletions).
	if err := stageChanges(wt); err != nil {
		return err
	}
	// Commit on the new branch.
	if err := commitChanges(wt, newBranchName); err != nil {
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

func GetBranchesWithStringName(prefix string) (map[string]string, error) {
	repo, _, _, _, err := getRepoContext()
	if err != nil {
		return nil, err
	}

	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	branches := make(map[string]string)
	now := time.Now()

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branchName := ref.Name().Short()

			if prefix == "" || strings.HasPrefix(branchName, prefix) {
				commit, err := repo.CommitObject(ref.Hash())
				if err != nil {
					return fmt.Errorf("failed to get commit for branch %s: %w", branchName, err)
				}

				timeSince := now.Sub(commit.Author.When)

				var timeStr string
				days := int(timeSince.Hours() / 24)
				if days > 0 {
					timeStr = fmt.Sprintf("%d days ago", days)
				} else {
					minutes := int(timeSince.Minutes())
					timeStr = fmt.Sprintf("%dmin ago", minutes)
				}

				branches[branchName] = timeStr
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error processing references: %w", err)
	}

	return branches, nil
}
