package gitx

import (
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

func setupTestRepo(t *testing.T) (string, func()) {
	remotePath, err := os.MkdirTemp("", "remote")
	require.NoError(t, err)
	_, err = git.PlainInit(remotePath, true)
	require.NoError(t, err)

	localPath, err := os.MkdirTemp("", "local")
	require.NoError(t, err)
	repo, err := git.PlainInit(localPath, false)
	require.NoError(t, err)

	originalWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(localPath))

	require.NoError(t, repo.Storer.SetReference(
		plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName("main")),
	))

	wt, err := repo.Worktree()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile("initial.txt", []byte("initial content"), 0o644))
	_, err = wt.Add("initial.txt")
	require.NoError(t, err)
	_, err = wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	_, err = repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{remotePath}})
	require.NoError(t, err)
	require.NoError(t, repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/main:refs/heads/main"},
	}))

	cleanup := func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Logf("failed to change back to original directory: %v", err)
		}
		_ = os.RemoveAll(remotePath)
		_ = os.RemoveAll(localPath)
	}

	return localPath, cleanup
}

func TestStashChangesToNewBranch(t *testing.T) {
	// Arrange
	localPath, cleanup := setupTestRepo(t)
	defer cleanup()
	newFilePath := filepath.Join(localPath, "new-feature.txt")
	require.NoError(t, os.WriteFile(newFilePath, []byte("work in progress"), 0o644))
	newBranchName := "feature/new-stuff"

	// Act
	err := StashChangesToNewBranch(newBranchName)

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

func TestGetBranchesWithStringName_FilterAndFormatting(t *testing.T) {
	// Arrange
	localPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	tenMinAgo := time.Now().Add(-10 * time.Minute)
	oneMinAgo := time.Now().Add(-1 * time.Minute)

	// create different branches
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

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))

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

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))

	// Act
	all, err := GetBranchesWithStringName("")
	require.NoError(t, err)
	only8stash, err := GetBranchesWithStringName("8stash/")
	require.NoError(t, err)

	// Assert
	assert.Contains(t, all, "main")                         // includes main
	assert.Contains(t, all, "8stash/xyz")                   // includes 8stash/xyz
	assert.Contains(t, all, "feature/abc")                  // includes feature/abc
	assert.Contains(t, all, "bugfix/one")                   // includes bugfix/one
	assert.NotEmpty(t, all["main"])                         // main has a formatted time string
	assert.Len(t, only8stash, 1)                            // prefix filter selects only 8stash/*
	assert.Equal(t, "2 days ago", only8stash["8stash/xyz"]) // time formatting for 2 days ago
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

func TestUpdateRepository_UpToDate(t *testing.T) {
	// Arrange
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Act
	err := UpdateRepository()

	// Assert
	require.NoError(t, err)
}

func TestUpdateRepository_FastForward(t *testing.T) {
	// Arrange
	localPath, cleanup := setupTestRepo(t)
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
		RefSpecs: []config.RefSpec{"refs/heads/main:refs/heads/main"},
	}))

	require.NoError(t, os.Chdir(localPath))

	// Act
	err = UpdateRepository()

	// Assert
	require.NoError(t, err) // fast-forward pull succeeds
}

func TestUpdateRepository_NonFastForward(t *testing.T) {
	// Arrange
	localPath, cleanup := setupTestRepo(t)
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
		RefSpecs: []config.RefSpec{"refs/heads/main:refs/heads/main"},
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

func TestPrepareRepository_CleanRepo_NoChanges_Error(t *testing.T) {
	// Arrange
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Act
	err := PrepareRepository()

	// Assert
	require.NoError(t, err) // clean repo passes PrepareRepository
}

func TestPrepareRepository_PropagatesUpdateError(t *testing.T) {
	// Arrange
	localPath, cleanup := setupTestRepo(t)
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
		RefSpecs: []config.RefSpec{"refs/heads/main:refs/heads/main"},
	}))

	require.NoError(t, os.Chdir(localPath))

	// Act
	err = PrepareRepository()

	// Assert
	require.Error(t, err)                            // PrepareRepository fails when UpdateRepository fails
	assert.ErrorContains(t, err, "non fast-forward") // error message indicates non fast-forward
}
