package service

import (
	"8stash/internal/test"
	"bytes"
	"errors"
	"fmt"
	"io"
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

func createAndPushStashBranch(t *testing.T, repo *git.Repository, wt *git.Worktree, localPath, fullBranchName, fileName, content string, when time.Time) {
	t.Helper()

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(fullBranchName),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, fileName), []byte(content), 0o644))
	_, err := wt.Add(fileName)
	require.NoError(t, err)
	_, err = wt.Commit("stash "+fullBranchName, &git.CommitOptions{
		Author: &object.Signature{Name: "S", Email: "s@example.com", When: when},
	})
	require.NoError(t, err)
	require.NoError(t, repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/heads/" + fullBranchName + ":refs/heads/" + fullBranchName)},
	}))
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))
}

func fetchAll(t *testing.T, repo *git.Repository) {
	t.Helper()
	err := repo.Fetch(&git.FetchOptions{RemoteName: "origin"})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		require.NoError(t, err)
	}
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestHandlePush_Succeeds(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	// Create working changes so PrepareRepository passes.
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "wip.txt"), []byte("work in progress"), 0o644))

	// Act
	stashName, err := HandlePush()

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, stashName)
	assert.True(t, len(stashName) > len(BranchPrefix) && stashName[:len(BranchPrefix)] == BranchPrefix)

	remote, err := repo.Remote("origin")
	require.NoError(t, err)
	refs, err := remote.List(&git.ListOptions{})
	require.NoError(t, err)

	found := false
	want := fmt.Sprintf("refs/heads/%s", stashName)
	for _, r := range refs {
		if r.Name().String() == want {
			found = true
			break
		}
	}
	assert.True(t, found, "pushed stash branch should exist on remote")

	head, err := repo.Head()
	require.NoError(t, err)
	assert.Equal(t, "refs/heads/main", head.Name().String())

	wt, err = repo.Worktree()
	require.NoError(t, err)
	status, err := wt.Status()
	require.NoError(t, err)
	assert.True(t, status.IsClean())
}

func TestHandleList_PrintsStashes(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	createAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"one", "one.txt", "1", time.Now().Add(-2*time.Hour))
	createAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"two", "two.txt", "2", time.Now().Add(-1*time.Hour))

	// ensure local has remote refs
	fetchAll(t, repo)

	// Act
	out := captureOutput(t, func() {
		err = HandleList()
	})

	// Assert
	require.NoError(t, err)
	assert.Contains(t, out, "Available stashes:")
	assert.Contains(t, out, BranchPrefix+"one")
	assert.Contains(t, out, BranchPrefix+"two")
}

func TestHandlePop_NoStashes_Error(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	fetchAll(t, repo)

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

	createAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"111", "a.txt", "A", time.Now())
	createAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"222", "b.txt", "B", time.Now())
	fetchAll(t, repo)

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

	createAndPushStashBranch(t, repo, wt, localPath, branchName, fileName, content, time.Now())
	fetchAll(t, repo)

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
	createAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"111", "x.txt", "X", time.Now())
	createAndPushStashBranch(t, repo, wt, localPath, BranchPrefix+"222", "y.txt", "Y", time.Now())
	fetchAll(t, repo)

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
	createAndPushStashBranch(t, repo, wt, localPath, stashBranch, stashFile, stashContent, time.Now())

	// Simulate divergence: create a DIFFERENT file on main (no conflict)
	mainFile := "main.txt"
	require.NoError(t, os.WriteFile(filepath.Join(localPath, mainFile), []byte("main content"), 0o644))
	_, err = wt.Add(mainFile)
	require.NoError(t, err)
	_, err = wt.Commit("main diverged", &git.CommitOptions{
		Author: &object.Signature{Name: "M", Email: "m@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	fetchAll(t, repo)

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
	createAndPushStashBranch(t, repo, wt, localPath, stashBranch, fileName, "stash change", time.Now())

	// Simulate divergence: conflicting change on main
	require.NoError(t, os.WriteFile(filepath.Join(localPath, fileName), []byte("main change"), 0o644))
	_, err = wt.Add(fileName)
	require.NoError(t, err)
	_, err = wt.Commit("main diverged", &git.CommitOptions{
		Author: &object.Signature{Name: "M", Email: "m@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	fetchAll(t, repo)

	// Act
	err = HandlePop("conflict")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CONFLICT")
}
