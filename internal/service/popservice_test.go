package service

import (
	"8stash/internal/test"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlePop_NoStashes_Error(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	test.FetchAll(t, repo)

	// Act
	err = HandlePop("0")

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "no pops found")
}

func TestHandlePop_Multiple_ZeroNumber_Error(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	test.CreateAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"111", "a.txt", "A", time.Now())
	test.CreateAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"222", "b.txt", "B", time.Now())
	test.FetchAll(t, repo)

	// Act
	err = HandlePop("0")

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "multiple pops found an no stash number given")
}

func TestHandlePop_Single_ZeroNumber_Succeeds(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	branchName := BranchPrefix + "solo"
	fileName := "solo.txt"
	content := "hello"

	test.CreateAndPushStashBranch(t, repo, wt, localPath, branchName, fileName, content, time.Now())
	test.FetchAll(t, repo)

	// Act
	err = HandlePop("0")

	// Assert
	require.NoError(t, err)

	// remote branch deleted
	remote, err := repo.Remote("origin")
	require.NoError(t, err)
	refs, err := remote.List(&git.ListOptions{})
	require.NoError(t, err)
	for _, r := range refs {
		assert.NotEqual(t, "refs/heads/"+branchName, r.Name().String())
	}

	// worktree has applied changes
	b, err := os.ReadFile(filepath.Join(localPath, fileName))
	require.NoError(t, err)
	assert.Equal(t, content, string(b))
	status, err := wt.Status()
	require.NoError(t, err)
	assert.False(t, status.IsClean())
}

func TestHandlePop_SelectByNumber_Succeeds(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	// two stashes
	test.CreateAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"111", "x.txt", "X", time.Now())
	test.CreateAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"222", "y.txt", "Y", time.Now())
	test.FetchAll(t, repo)

	// Act
	err = HandlePop("111")

	// Assert we do not assert that HandlePop has no Error because this is supposed to happen when no brach is found after pop
	remote, err := repo.Remote("origin")
	require.NoError(t, err)
	refs, err := remote.List(&git.ListOptions{})
	require.NoError(t, err)

	var has111, has222 bool
	for _, r := range refs {
		switch r.Name().String() {
		case "refs/heads/" + BranchPrefix + "111":
			has111 = true
		case "refs/heads/" + BranchPrefix + "222":
			has222 = true
		}
	}
	assert.False(t, has111, "selected stash should be deleted")
	assert.True(t, has222, "other stash should remain")

	// applied content from 111 present
	b, err := os.ReadFile(filepath.Join(localPath, "x.txt"))
	require.NoError(t, err)
	assert.Equal(t, "X", string(b))
}

func TestHandlePop_DivergedBranches_TriggersMerge(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	// Create and push a stash branch with one file
	stashBranch := BranchPrefix + "diverge"
	stashFile := "stash.txt"
	stashContent := "stash content"
	test.CreateAndPushStashBranch(t, repo, wt, localPath, stashBranch, stashFile, stashContent, time.Now())

	// Simulate divergence: create a DIFFERENT file on main (no conflict)
	mainFile := "main.txt"
	require.NoError(t, os.WriteFile(filepath.Join(localPath, mainFile), []byte("main content"), 0o644))
	_, err = wt.Add(mainFile)
	require.NoError(t, err)
	_, err = wt.Commit("main diverged", &git.CommitOptions{
		Author: &object.Signature{Name: "M", Email: "m@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	test.FetchAll(t, repo)

	// Act
	err = HandlePop("diverge")

	// Assert - should succeed with no error
	require.NoError(t, err)

	// Both files should exist after successful merge
	stashBytes, err := os.ReadFile(filepath.Join(localPath, stashFile))
	require.NoError(t, err)
	assert.Equal(t, stashContent, string(stashBytes))

	mainBytes, err := os.ReadFile(filepath.Join(localPath, mainFile))
	require.NoError(t, err)
	assert.Equal(t, "main content", string(mainBytes))
}

func TestHandlePop_DivergedBranches_MergeConflict(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	stashBranch := BranchPrefix + "conflict"
	fileName := "conflict.txt"
	test.CreateAndPushStashBranch(t, repo, wt, localPath, stashBranch, fileName, "stash change", time.Now())

	// Simulate divergence: conflicting change on main
	require.NoError(t, os.WriteFile(filepath.Join(localPath, fileName), []byte("main change"), 0o644))
	_, err = wt.Add(fileName)
	require.NoError(t, err)
	_, err = wt.Commit("main diverged", &git.CommitOptions{
		Author: &object.Signature{Name: "M", Email: "m@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	test.FetchAll(t, repo)

	// Act
	err = HandlePop("conflict")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CONFLICT")
}
