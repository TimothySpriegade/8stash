package gitx

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"8stash/internal/test"
)

func TestStashChangesToNewBranch(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()
	newFilePath := filepath.Join(localPath, "new-feature.txt")
	require.NoError(t, os.WriteFile(newFilePath, []byte("work in progress"), 0o644))
	newBranchName := "feature/new-stuff"

	// Act
	err := StashChangesToNewBranch(newBranchName, "")

	// Assert
	require.NoError(t, err) // operation succeeds without error

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)

	head, err := repo.Head()
	require.NoError(t, err)
	assert.Equal(t, "refs/heads/main", head.Name().String()) //  back on main

	wt, err := repo.Worktree()
	require.NoError(t, err)
	status, err := wt.Status()
	require.NoError(t, err)
	assert.True(t, status.IsClean()) // working directory is clean

	remote, err := repo.Remote("origin")
	require.NoError(t, err)
	refs, err := remote.List(&git.ListOptions{})
	require.NoError(t, err)

	found := false
	expectedRef := fmt.Sprintf("refs/heads/%s", newBranchName)
	for _, ref := range refs {
		if ref.Name().String() == expectedRef {
			found = true
			break
		}
	}
	assert.True(t, found) // new branch exists on the remote
}

func TestStashChangesToNewBranch_EmptyName_Error(t *testing.T) {
	// Arrange
	_, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Act
	err := StashChangesToNewBranch("", "")

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "branch name must not be empty")
}

func TestStashChangesToNewBranch_TargetEqualsCurrent_Error(t *testing.T) {
	// Arrange
	_, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Act
	err := StashChangesToNewBranch("main", "")

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "target branch equals current branch")
}

func TestStashChangesToNewBranch_TargetAlreadyExists_Error(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	exists := "feature/exist"
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(exists),
		Create: true,
		Keep:   true,
	}))
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))

	// Act
	err = StashChangesToNewBranch(exists, "")

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "already exists")
}

func TestStashChangesToNewBranch_WithCustomMessage_UsesCustomMessage(t *testing.T) {
    // Arrange
    localPath, cleanup := test.SetupTestRepo(t)
    defer cleanup()
    newFilePath := filepath.Join(localPath, "feature.txt")
    require.NoError(t, os.WriteFile(newFilePath, []byte("new feature"), 0o644))
    newBranchName := "feature/custom-msg"
    customMessage := "WIP: implementing new login flow"

    // Act
    err := StashChangesToNewBranch(newBranchName, customMessage)

    // Assert
    require.NoError(t, err)
    repo, err := git.PlainOpen(localPath)
    require.NoError(t, err)
    ref, err := repo.Reference(plumbing.NewBranchReferenceName(newBranchName), true)
    require.NoError(t, err)
    commit, err := repo.CommitObject(ref.Hash())
    require.NoError(t, err)
    assert.Equal(t, customMessage, commit.Message)
}

func TestStashChangesToNewBranch_WithEmptyMessage_UsesDefaultMessage(t *testing.T) {
    // Arrange
    localPath, cleanup := test.SetupTestRepo(t)
    defer cleanup()
    newFilePath := filepath.Join(localPath, "feature.txt")
    require.NoError(t, os.WriteFile(newFilePath, []byte("new feature"), 0o644))
    newBranchName := "feature/default-msg"

    // Act
    err := StashChangesToNewBranch(newBranchName, "")

    // Assert
    require.NoError(t, err)
    repo, err := git.PlainOpen(localPath)
    require.NoError(t, err)
    ref, err := repo.Reference(plumbing.NewBranchReferenceName(newBranchName), true)
    require.NoError(t, err)
    commit, err := repo.CommitObject(ref.Hash())
    require.NoError(t, err)
    expectedDefaultMsg := fmt.Sprintf("move local changes to branch %s", newBranchName)
    assert.Equal(t, expectedDefaultMsg, commit.Message)
}