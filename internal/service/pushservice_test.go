package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"8stash/internal/constants"
	"8stash/internal/test"
)

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
	assert.True(t, len(stashName) > len(constants.BranchPrefix) && stashName[:len(constants.BranchPrefix)] == constants.BranchPrefix)

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
