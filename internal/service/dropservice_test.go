package service

import (
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"8stash/internal/test"
	"8stash/internal/config"
)

func TestHandleDrop_Succeeds(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	branchToDrop := "droppable"
	fullBranchName := config.BranchPrefix + branchToDrop
	test.CreateAndPushStashBranch(t, repo, wt, localPath, fullBranchName, "drop.txt", "content", time.Now())
	test.FetchAll(t, repo)

	// Act
	err = HandleDrop(branchToDrop)

	// Assert
	require.NoError(t, err)

	remote, err := repo.Remote("origin")
	require.NoError(t, err)
	refs, err := remote.List(&git.ListOptions{})
	require.NoError(t, err)

	found := false
	want := "refs/heads/" + fullBranchName
	for _, r := range refs {
		if r.Name().String() == want {
			found = true
			break
		}
	}
	assert.False(t, found, "dropped stash branch should not exist on remote")
}

func TestHandleDrop_BranchNotFound_NoError(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	_, err := git.PlainOpen(localPath)
	require.NoError(t, err)

	// Act
	err = HandleDrop("nonexistent")

	// Assert
	require.NoError(t, err)
}
