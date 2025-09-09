package gitx

import (
	"8stash/internal/test"
	"fmt"
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

func TestDeleteBranch_Succeeds_LocalAndRemote(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	branch := "feature/temp"

	// Create branch with a commit and push it
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "temp.txt"), []byte("x"), 0o644))
	_, err = wt.Add("temp.txt")
	require.NoError(t, err)
	_, err = wt.Commit("temp", &git.CommitOptions{
		Author: &object.Signature{Name: "X", Email: "x@example.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, repo.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/" + branch + ":refs/heads/" + branch)},
	}))

	// Switch back to main
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))

	// Act
	err = DeleteBranch(branch)

	// Assert
	require.NoError(t, err)

	// Local ref removed
	_, err = repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	assert.Error(t, err)

	// Remote ref removed
	remote, err := repo.Remote("origin")
	require.NoError(t, err)
	refs, err := remote.List(&git.ListOptions{})
	require.NoError(t, err)
	for _, r := range refs {
		assert.NotEqual(t, fmt.Sprintf("refs/heads/%s", branch), r.Name().String())
	}
}

func TestDeleteBranch_CurrentBranch_Error(t *testing.T) {
	// Arrange
	_, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Act
	err := DeleteBranch("main")

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot delete current branch")
}

func TestDeleteBranch_EmptyName_Error(t *testing.T) {
	// Arrange
	_, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Act
	err := DeleteBranch("")

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "branch name must not be empty")
}
