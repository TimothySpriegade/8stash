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

func MergeStashIntoCurrentBranch(branchName string) error {
	repo, wt, currentBranch, remote, err := getRepoContext()
	if err != nil {
		return err
	}
	if strings.TrimSpace(branchName) == "" {
		return fmt.Errorf(branchNameMustNotEmptyErrorMsg)
	}

	cands, err := findRemoteCandidates(repo, branchName)
	if err != nil {
		return err
	}
	if len(cands) == 0 {
		return fmt.Errorf("no remote branch %q found", branchName)
	}

	var target *plumbing.Reference
	prefer := plumbing.ReferenceName("refs/remotes/" + remote + "/" + strings.TrimPrefix(branchName, remote+"/"))
	for _, r := range cands {
		if r.Name() == prefer {
			target = r
			break
		}
	}
	if target == nil {
		for _, r := range cands {
			if strings.HasSuffix(r.Name().Short(), "/HEAD") {
				continue
			}
			target = r
			break
		}
	}
	if target == nil {
		return fmt.Errorf("no suitable remote branch candidate for %q", branchName)
	}

	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("HEAD: %w", err)
	}
	originalHeadHash := headRef.Hash()

	ok, err := isAncestor(repo, headRef.Hash(), target.Hash())
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("non fast-forward merge required: %s -> %s",
			headRef.Name().Short(), target.Name().Short())
	}

	brName := plumbing.NewBranchReferenceName(currentBranch)
	if err := repo.Storer.SetReference(plumbing.NewHashReference(brName, target.Hash())); err != nil {
		return fmt.Errorf("update branch ref: %w", err)
	}
	if err := wt.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: target.Hash(),
	}); err != nil {
		return fmt.Errorf("reset worktree: %w", err)
	}

	if err := wt.Reset(&git.ResetOptions{
		Mode:   git.MixedReset,
		Commit: originalHeadHash,
	}); err != nil {
		return fmt.Errorf("mixed reset to original head: %w", err)
	}

	return nil
}

func isAncestor(repo *git.Repository, ancestor, descendant plumbing.Hash) (bool, error) {
	if ancestor == descendant {
		return true, nil
	}

	seen := make(map[plumbing.Hash]struct{})
	queue := []plumbing.Hash{descendant}

	for len(queue) > 0 {
		h := queue[0]
		queue = queue[1:]

		if h == ancestor {
			return true, nil
		}
		if _, ok := seen[h]; ok {
			continue
		}
		seen[h] = struct{}{}

		c, err := repo.CommitObject(h)
		if err != nil {
			return false, err
		}
		for _, ph := range c.ParentHashes {
			if ph == ancestor {
				return true, nil
			}
			if _, ok := seen[ph]; !ok {
				queue = append(queue, ph)
			}
		}
	}
	return false, nil
}

func findRemoteCandidates(repo *git.Repository, branchName string) ([]*plumbing.Reference, error) {
	var out []*plumbing.Reference

	if strings.Contains(branchName, "/") {
		exact := plumbing.ReferenceName("refs/remotes/" + branchName)
		if ref, err := repo.Reference(exact, true); err == nil {
			out = append(out, ref)
			return out, nil
		}
	}

	iter, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("list references: %w", err)
	}
	defer iter.Close()

	wantSuffix := "/" + strings.TrimPrefix(branchName, "/")
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		if !ref.Name().IsRemote() {
			return nil
		}
		if strings.HasSuffix(ref.Name().Short(), wantSuffix) {
			out = append(out, ref)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("iterate references: %w", err)
	}
	return out, nil
}

func DeleteBranch(branchName string) error {
	repo, _, _, remoteName, err := getRepoContext() // Assumes getRepoContext exists and returns these values
	if err != nil {
		return err
	}

	if strings.TrimSpace(branchName) == "" {
		return fmt.Errorf(branchNameMustNotEmptyErrorMsg)
	}

	// Delete the local branch
	fmt.Printf("trying to delete branch %s locally\n", branchName)
	localRefName := plumbing.NewBranchReferenceName(branchName)

	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("could not read HEAD: %w", err)
	}
	if headRef.Name() == localRefName {
		return fmt.Errorf("cannot delete current branch %q. Please switch to another branch first", branchName)
	}

	err = repo.Storer.RemoveReference(localRefName)
	if err != nil && err != plumbing.ErrReferenceNotFound {
		return fmt.Errorf("failed to delete local branch %q: %w", branchName, err)
	}
	if err == nil {
		fmt.Printf("Local branch '%s' was deleted successfully.", branchName)
	} else {
		fmt.Printf("Local branch '%s' not found or already deleted.", branchName)
	}

	// Delete the remote branch
	fmt.Printf("trying to delete branch %s on remote\n", branchName)
	remoteRefSpec := config.RefSpec(":" + localRefName.String())

	pushOptions := &git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   []config.RefSpec{remoteRefSpec},
	}

	fmt.Printf("Attempting to delete remote branch '%s' on '%s'...", branchName, remoteName)

	err = repo.Push(pushOptions)

	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to delete remote branch: %w", err)
	}

	fmt.Printf("Remote branch '%s' on '%s' deleted successfully or was not present.", branchName, remoteName)

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
		return fmt.Errorf(branchNameMustNotEmptyErrorMsg)
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
	defer refs.Close()

	branches := make(map[string]string)
	now := time.Now()

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if !ref.Name().IsRemote() || ref.Type() != plumbing.HashReference {
			return nil
		}

		short := ref.Name().Short()
		parts := strings.SplitN(short, "/", 2)
		if len(parts) != 2 {
			return nil
		}
		remoteName, branchName := parts[0], parts[1]
		if remoteName != "origin" {
			return nil
		}

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
				hours := int(timeSince.Hours())
				if hours > 0 {
					timeStr = fmt.Sprintf("%dh ago", hours)
				} else {
					minutes := int(timeSince.Minutes())
					timeStr = fmt.Sprintf("%dmin ago", minutes)
				}
			}

			branches[branchName] = timeStr
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error processing references: %w", err)
	}

	return branches, nil
}
