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

func TestGetBranchesWithStringName_FilterAndFormatting(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	pushBranch := func(name string) {
		require.NoError(t, repo.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs:   []config.RefSpec{config.RefSpec("refs/heads/" + name + ":refs/heads/" + name)},
		}))
	}

	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	tenMinAgo := time.Now().Add(-10 * time.Minute)
	oneMinAgo := time.Now().Add(-1 * time.Minute)

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("8stash/xyz"),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "feat_xyz.txt"), []byte("feat xyz"), 0o644))
	_, err = wt.Add("feat_xyz.txt")
	require.NoError(t, err)
	_, err = wt.Commit("feat xyz", &git.CommitOptions{
		Author: &object.Signature{Name: "T", Email: "t@example.com", When: twoDaysAgo},
	})
	require.NoError(t, err)
	pushBranch("8stash/xyz")

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))

	// feature/abc
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("feature/abc"),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "feature_abc.txt"), []byte("feature abc"), 0o644))
	_, err = wt.Add("feature_abc.txt")
	require.NoError(t, err)
	_, err = wt.Commit("feature abc", &git.CommitOptions{
		Author: &object.Signature{Name: "T", Email: "t@example.com", When: tenMinAgo},
	})
	require.NoError(t, err)
	pushBranch("feature/abc")

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("bugfix/one"),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "bugfix_one.txt"), []byte("bugfix one"), 0o644))
	_, err = wt.Add("bugfix_one.txt")
	require.NoError(t, err)
	_, err = wt.Commit("bugfix one", &git.CommitOptions{
		Author: &object.Signature{Name: "T", Email: "t@example.com", When: oneMinAgo},
	})
	require.NoError(t, err)
	pushBranch("bugfix/one")

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))

	if err := repo.Fetch(&git.FetchOptions{RemoteName: "origin"}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		require.NoError(t, err)
	}

	// Act
	all, err := GetBranchesWithStringName("")
	require.NoError(t, err)
	only8stash, err := GetBranchesWithStringName("8stash/")
	require.NoError(t, err)

	// Assert
	assert.Contains(t, all, "main")
	assert.Contains(t, all, "8stash/xyz")
	assert.Contains(t, all, "feature/abc")
	assert.Contains(t, all, "bugfix/one")
	assert.NotEmpty(t, all["main"])
	assert.Len(t, only8stash, 1)
	assert.Equal(t, "2 days ago", only8stash["8stash/xyz"])
}

func TestGetBranchesWithStringName_NotARepo(t *testing.T) {
	// Arrange
	orig, err := os.Getwd()
	require.NoError(t, err)
	tmp, err := os.MkdirTemp("", "not-a-repo")
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(orig)
		_ = os.RemoveAll(tmp)
	}()
	require.NoError(t, os.Chdir(tmp))

	// Act
	_, err = GetBranchesWithStringName("")

	// Assert
	require.Error(t, err)
}
