package gitx

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

func findBestRemoteCandidate(candidates []*plumbing.Reference, remote, branchName string) *plumbing.Reference {
	var fallback *plumbing.Reference
	prefer := plumbing.ReferenceName("refs/remotes/" + remote + "/" + strings.TrimPrefix(branchName, remote+"/"))

	for _, r := range candidates {
		if r.Name() == prefer {
			return r
		}
		if fallback == nil && !strings.HasSuffix(r.Name().Short(), "/HEAD") {
			fallback = r
		}
	}
	return fallback
}

func MergeStashIntoCurrentBranch(branchName string) error {
	repo, wt, currentBranch, remote, err := getRepoContext()
	if err != nil {
		return err
	}
	if strings.TrimSpace(branchName) == "" {
		return fmt.Errorf(branchNameMustNotEmptyErrorMsg)
	}

	candidates, err := findRemoteCandidates(repo, branchName)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no remote branch %q found", branchName)
	}

	target := findBestRemoteCandidate(candidates, remote, branchName)
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
	if err := wt.Reset(&git.ResetOptions{Mode: git.HardReset, Commit: target.Hash()}); err != nil {
		return fmt.Errorf("reset worktree: %w", err)
	}
	if err := wt.Reset(&git.ResetOptions{Mode: git.MixedReset, Commit: originalHeadHash}); err != nil {
		return fmt.Errorf("mixed reset to original head: %w", err)
	}

	return nil
}

func processCommitNode(repo *git.Repository, hash plumbing.Hash, queue *[]plumbing.Hash, seen map[plumbing.Hash]struct{}) error {
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return err
	}

	for _, parentHash := range commit.ParentHashes {
		if _, ok := seen[parentHash]; !ok {
			*queue = append(*queue, parentHash)
		}
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
		currentHash := queue[0]
		queue = queue[1:]
		if currentHash == ancestor {
			return true, nil
		}
		if _, ok := seen[currentHash]; ok {
			continue
		}
		seen[currentHash] = struct{}{}
		if err := processCommitNode(repo, currentHash, &queue, seen); err != nil {
			return false, err
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
