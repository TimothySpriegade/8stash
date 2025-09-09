package gitx

import (
	"8stash/internal/test"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeStashIntoCurrentBranch_FastForward_AppliesWorktreeChanges(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	// Create remote branch from main with an extra commit
	branchName := "8stash/xyz"
	fileName := "merge-stash.txt"

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, fileName), []byte("remote work"), 0o644))
	_, err = wt.Add(fileName)
	require.NoError(t, err)
	_, err = wt.Commit("remote change", &git.CommitOptions{
		Author: &object.Signature{Name: "R", Email: "r@example.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, repo.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/" + branchName + ":refs/heads/" + branchName)},
	}))

	// Switch back to main
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))
	headBefore, err := repo.Head()
	require.NoError(t, err)
	origHash := headBefore.Hash()

	// Act
	err = MergeStashIntoCurrentBranch(branchName)

	// Assert
	require.NoError(t, err)

	headAfter, err := repo.Head()
	require.NoError(t, err)
	assert.Equal(t, "refs/heads/main", headAfter.Name().String())
	// HEAD should be at the original main commit (changes applied to worktree)
	assert.Equal(t, origHash, headAfter.Hash())

	wt, err = repo.Worktree()
	require.NoError(t, err)
	status, err := wt.Status()
	require.NoError(t, err)
	assert.False(t, status.IsClean())
	b, err := os.ReadFile(filepath.Join(localPath, fileName))
	require.NoError(t, err)
	assert.Equal(t, "remote work", string(b))
}

func TestMergeStashIntoCurrentBranch_NonFastForward_Error(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	branchName := "8stash/xyz"

	// Create and push the remote branch off main
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "remote.txt"), []byte("remote"), 0o644))
	_, err = wt.Add("remote.txt")
	require.NoError(t, err)
	_, err = wt.Commit("remote change", &git.CommitOptions{
		Author: &object.Signature{Name: "R", Email: "r@example.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, repo.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/" + branchName + ":refs/heads/" + branchName)},
	}))

	// Diverge local main with a new local commit
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "local.txt"), []byte("local"), 0o644))
	_, err = wt.Add("local.txt")
	require.NoError(t, err)
	_, err = wt.Commit("local change", &git.CommitOptions{
		Author: &object.Signature{Name: "L", Email: "l@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	// Act
	err = MergeStashIntoCurrentBranch(branchName)

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "non fast-forward")
}

func TestApplyDivergedMerge_Succeeds_NoConflicts(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	branchName := "feature/merge-ok"
	fileName := "remote_ok.txt"

	// Create and push remote branch from main with a non-conflicting change
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, fileName), []byte("remote content"), 0o644))
	_, err = wt.Add(fileName)
	require.NoError(t, err)
	_, err = wt.Commit("remote non-conflicting change", &git.CommitOptions{
		Author: &object.Signature{Name: "R", Email: "r@example.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, repo.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/" + branchName + ":refs/heads/" + branchName)},
	}))

	// Switch back to main and ensure we have the remote-tracking ref locally
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))
	if err := repo.Fetch(&git.FetchOptions{RemoteName: "origin"}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		require.NoError(t, err)
	}

	headBefore, err := repo.Head()
	require.NoError(t, err)
	origHash := headBefore.Hash()

	// Act
	err = ApplyDivergedMerge(branchName)

	// Assert
	require.NoError(t, err)

	headAfter, err := repo.Head()
	require.NoError(t, err)
	assert.Equal(t, "refs/heads/main", headAfter.Name().String())
	assert.Equal(t, origHash, headAfter.Hash())

	wt, err = repo.Worktree()
	require.NoError(t, err)
	status, err := wt.Status()
	require.NoError(t, err)
	assert.False(t, status.IsClean())

	b, err := os.ReadFile(filepath.Join(localPath, fileName))
	require.NoError(t, err)
	assert.Equal(t, "remote content", string(b))
}

func TestApplyDivergedMerge_WithConflicts_ReturnsError(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	branchName := "feature/conflict"
	fileName := "same.txt"

	// Create and push remote branch with a change
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, fileName), []byte("remote change\n"), 0o644))
	_, err = wt.Add(fileName)
	require.NoError(t, err)
	_, err = wt.Commit("remote conflicting change", &git.CommitOptions{
		Author: &object.Signature{Name: "R", Email: "r@example.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, repo.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/" + branchName + ":refs/heads/" + branchName)},
	}))

	// Switch back to main and create a conflicting local change
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, fileName), []byte("local change\n"), 0o644))
	_, err = wt.Add(fileName)
	require.NoError(t, err)
	_, err = wt.Commit("local conflicting change", &git.CommitOptions{
		Author: &object.Signature{Name: "L", Email: "l@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	// Ensure remote-tracking ref exists locally
	if err := repo.Fetch(&git.FetchOptions{RemoteName: "origin"}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		require.NoError(t, err)
	}

	// Act
	err = ApplyDivergedMerge(branchName)

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "automatic merge failed")
}

func TestApplyDivergedMerge_NoRemoteCandidate_Error(t *testing.T) {
	// Arrange
	_, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Act
	err := ApplyDivergedMerge("does/not/exist")

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "no suitable remote branch candidate")
}
