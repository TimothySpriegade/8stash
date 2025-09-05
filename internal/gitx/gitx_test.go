package gitx

import (
	"8stash/internal/test"
	"errors"
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

func TestStashChangesToNewBranch(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
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

func TestStashChangesToNewBranch_EmptyName_Error(t *testing.T) {
	// Arrange
	_, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Act
	err := StashChangesToNewBranch("")

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "branch name must not be empty")
}

func TestStashChangesToNewBranch_TargetEqualsCurrent_Error(t *testing.T) {
	// Arrange
	_, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	// Act
	err := StashChangesToNewBranch("main")

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
	err = StashChangesToNewBranch(exists)

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "already exists")
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
