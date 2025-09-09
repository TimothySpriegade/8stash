package gitx

import (
	"8stash/internal/test"
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

func TestUpdateRepository_UpToDate(t *testing.T) {
	// Arrange
	_, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Act
	err := UpdateRepository()

	// Assert
	require.NoError(t, err)
}

func TestUpdateRepository_FastForward(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Second clone on main that pushes a new commit to origin.
	repo1, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	cfg, err := repo1.Config()
	require.NoError(t, err)
	originURL := cfg.Remotes["origin"].URLs[0]

	otherPath, err := os.MkdirTemp("", "other")
	require.NoError(t, err)
	defer os.RemoveAll(otherPath)

	repo2, err := git.PlainClone(otherPath, &git.CloneOptions{
		URL:           originURL,
		ReferenceName: plumbing.NewBranchReferenceName("main"),
		SingleBranch:  true,
	})
	require.NoError(t, err)
	wt2, err := repo2.Worktree()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(otherPath, "ff.txt"), []byte("ff"), 0o644))
	_, err = wt2.Add("ff.txt")
	require.NoError(t, err)
	_, err = wt2.Commit("remote ff", &git.CommitOptions{
		Author: &object.Signature{Name: "R2", Email: "r2@example.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, repo2.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/main:refs/heads/main")},
	}))

	require.NoError(t, os.Chdir(localPath))

	// Act
	err = UpdateRepository()

	// Assert
	require.NoError(t, err) // fast-forward pull succeeds
}

func TestUpdateRepository_NonFastForward(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Local commit (not pushed) to create divergence.
	repo1, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt1, err := repo1.Worktree()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "local.txt"), []byte("local"), 0o644))
	_, err = wt1.Add("local.txt")
	require.NoError(t, err)
	_, err = wt1.Commit("local diverge", &git.CommitOptions{
		Author: &object.Signature{Name: "L1", Email: "l1@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	// Remote gets an independent commit on main.
	cfg, err := repo1.Config()
	require.NoError(t, err)
	originURL := cfg.Remotes["origin"].URLs[0]

	otherPath, err := os.MkdirTemp("", "other")
	require.NoError(t, err)
	defer os.RemoveAll(otherPath)

	repo2, err := git.PlainClone(otherPath, &git.CloneOptions{
		URL:           originURL,
		ReferenceName: plumbing.NewBranchReferenceName("main"),
		SingleBranch:  true,
	})
	require.NoError(t, err)
	wt2, err := repo2.Worktree()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(otherPath, "remote.txt"), []byte("remote"), 0o644))
	_, err = wt2.Add("remote.txt")
	require.NoError(t, err)
	_, err = wt2.Commit("remote diverge", &git.CommitOptions{
		Author: &object.Signature{Name: "R2", Email: "r2@example.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, repo2.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/main:refs/heads/main")},
	}))

	require.NoError(t, os.Chdir(localPath))

	// Act
	err = UpdateRepository()

	// Assert
	require.Error(t, err)                            // pull fails on divergence
	assert.ErrorContains(t, err, "non fast-forward") // specific non fast-forward error is returned
}

func TestPrepareRepository_NotARepo(t *testing.T) {
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
	err = PrepareRepository()

	// Assert
	require.Error(t, err) // fails validation when not a repository
}

func TestPrepareRepository_CleanRepo_NoChanges(t *testing.T) {
	// Arrange
	_, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Act
	err := PrepareRepository()

	// Assert
	require.Error(t, err)
}

func TestPrepareRepository_PropagatesUpdateError(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Local commit (not pushed) to diverge.
	repo1, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt1, err := repo1.Worktree()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "local-div.txt"), []byte("x"), 0o644))
	_, err = wt1.Add("local-div.txt")
	require.NoError(t, err)
	_, err = wt1.Commit("local diverge", &git.CommitOptions{
		Author: &object.Signature{Name: "L1", Email: "l1@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	// Remote gets an independent commit on main.
	cfg, err := repo1.Config()
	require.NoError(t, err)
	originURL := cfg.Remotes["origin"].URLs[0]

	otherPath, err := os.MkdirTemp("", "other")
	require.NoError(t, err)
	defer os.RemoveAll(otherPath)

	repo2, err := git.PlainClone(otherPath, &git.CloneOptions{
		URL:           originURL,
		ReferenceName: plumbing.NewBranchReferenceName("main"),
		SingleBranch:  true,
	})
	require.NoError(t, err)
	wt2, err := repo2.Worktree()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(otherPath, "remote-div.txt"), []byte("y"), 0o644))
	_, err = wt2.Add("remote-div.txt")
	require.NoError(t, err)
	_, err = wt2.Commit("remote diverge", &git.CommitOptions{
		Author: &object.Signature{Name: "R2", Email: "r2@example.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, repo2.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/main:refs/heads/main")},
	}))

	require.NoError(t, os.Chdir(localPath))

	// Act
	err = PrepareRepository()

	// Assert
	require.Error(t, err)                            // PrepareRepository fails when UpdateRepository fails
	assert.ErrorContains(t, err, "non fast-forward") // error message indicates non fast-forward
}

func TestPrepareRepository_WithChanges_Succeeds(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	require.NoError(t, os.WriteFile(filepath.Join(localPath, "dirty.txt"), []byte("x"), 0o644))

	// Act
	err := PrepareRepository()

	// Assert
	require.NoError(t, err)
}
