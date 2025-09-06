package gitx

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

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
