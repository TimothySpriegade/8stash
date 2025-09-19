package service

import (
	"path/filepath"
	"testing"
	"time"
	"os"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"8stash/internal/config"
	"8stash/internal/test"
)

func TestHandleCleanup_NoStashes_Succeeds(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	test.FetchAll(t, repo)

	// Act
	err = HandleCleanup()

	// Assert
	require.NoError(t, err)
}

func TestHandleCleanup_NoOldStashes_NoDeletion(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	newer := time.Now().Add(-10 * 24 * time.Hour)
	test.CreateAndPushStashBranch(t, repo, wt, localPath, config.BranchPrefix+"new1", "n1.txt", "N1", newer)
	test.CreateAndPushStashBranch(t, repo, wt, localPath, config.BranchPrefix+"new2", "n2.txt", "N2", newer)
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")}))
	test.FetchAll(t, repo)

	// Act
	err = HandleCleanup()

	// Assert
	require.NoError(t, err)

	remote, err := repo.Remote("origin")
	require.NoError(t, err)
	refs, err := remote.List(&git.ListOptions{})
	require.NoError(t, err)

	has := func(refs []*plumbing.Reference, full string) bool {
		for _, r := range refs {
			if r.Name().String() == full {
				return true
			}
		}
		return false
	}

	assert.True(t, has(refs, "refs/heads/"+config.BranchPrefix+"new1"))
	assert.True(t, has(refs, "refs/heads/"+config.BranchPrefix+"new2"))
}

func TestHandleCleanup_DeletesOnlyOldStashes(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	config.SkipConfirmations = true
	defer cleanup()
	defer func() { config.SkipConfirmations = false }()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	oldWhen := time.Now().Add(-30 * 24 * time.Hour)
	newWhen := time.Now().Add(-5 * 24 * time.Hour)

	old1 := config.BranchPrefix + "old1"
	old2 := config.BranchPrefix + "old2"
	new1 := config.BranchPrefix + "new1"

	test.CreateAndPushStashBranch(t, repo, wt, localPath, old1, "o1.txt", "O1", oldWhen)
	test.CreateAndPushStashBranch(t, repo, wt, localPath, old2, "o2.txt", "O2", oldWhen)
	test.CreateAndPushStashBranch(t, repo, wt, localPath, new1, "n1.txt", "N1", newWhen)

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")}))
	test.FetchAll(t, repo)

	// Act
	err = HandleCleanup()

	// Assert
	require.NoError(t, err)

	remote, err := repo.Remote("origin")
	require.NoError(t, err)
	refs, err := remote.List(&git.ListOptions{})
	require.NoError(t, err)

	full := func(name string) string { return filepath.ToSlash("refs/heads/" + name) }
	var hasOld1, hasOld2, hasNew1 bool
	for _, r := range refs {
		switch r.Name().String() {
		case full(old1):
			hasOld1 = true
		case full(old2):
			hasOld2 = true
		case full(new1):
			hasNew1 = true
		}
	}
	assert.False(t, hasOld1, "old1 should be deleted")
	assert.False(t, hasOld2, "old2 should be deleted")
	assert.True(t, hasNew1, "new1 should remain")
}

func TestHandleCleanup_DeleteOldCurrentBranch_ReturnsError(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	config.SkipConfirmations = true
	defer cleanup()
	defer func() { config.SkipConfirmations = false }()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	old := config.BranchPrefix + "current-old"
	oldWhen := time.Now().Add(-30 * 24 * time.Hour)
	test.CreateAndPushStashBranch(t, repo, wt, localPath, old, "c.txt", "C", oldWhen)

	// Make the old stash the current branch to trigger delete error
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(old),
	}))
	test.FetchAll(t, repo)

	// Act
	err = HandleCleanup()

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "drop branch "+old)
	assert.ErrorContains(t, err, "cannot delete current branch")
}

func TestFilterBranches_FiltersOnlyDaysAgoAndOlderOrEqual(t *testing.T) {
	branches := map[string]string{
		"b1": "30 days ago", // keep (== limit)
		"b2": "29 days ago", // drop (< limit)
		"b3": "45 days ago", // keep (> limit)
		"b4": "1 day ago",   // drop (not 'days')
		"b5": "3 hours ago", // drop (not 'days')
		"b6": "invalid",     // drop (parse error)
	}
	out := filterBranches(branches, config.CleanUpTimeInDays)

	require.Len(t, out, 2)
	assert.Equal(t, "30 days ago", out["b1"])
	assert.Equal(t, "45 days ago", out["b3"])
	assert.NotContains(t, out, "b2")
	assert.NotContains(t, out, "b4")
	assert.NotContains(t, out, "b5")
	assert.NotContains(t, out, "b6")
}

func TestHandleCleanup_WithConfirmation_DeletesBranch(t *testing.T) {
    // Arrange
    // Mock stdin to simulate user typing "Y"
    oldStdin := os.Stdin
    r, w, err := os.Pipe()
    require.NoError(t, err)
    os.Stdin = r
    _, err = w.Write([]byte("Y\n"))
    require.NoError(t, err)
    w.Close()
    localPath, cleanup := test.SetupTestRepo(t)
    defer cleanup()
    defer func() { os.Stdin = oldStdin }()

    repo, err := git.PlainOpen(localPath)
    require.NoError(t, err)
    wt, err := repo.Worktree()
    require.NoError(t, err)

    oldWhen := time.Now().Add(-30 * 24 * time.Hour)
    oldBranchName := config.BranchPrefix + "old-to-delete"
    test.CreateAndPushStashBranch(t, repo, wt, localPath, oldBranchName, "old.txt", "old", oldWhen)
    require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")}))
    test.FetchAll(t, repo)

    // Act
    err = HandleCleanup()

    // Assert
    require.NoError(t, err)
    remote, err := repo.Remote("origin")
    require.NoError(t, err)
    refs, err := remote.List(&git.ListOptions{})
    require.NoError(t, err)

    var branchFound bool
    for _, ref := range refs {
        if ref.Name().String() == "refs/heads/"+oldBranchName {
            branchFound = true
            break
        }
    }
    assert.False(t, branchFound, "Expected old branch to be deleted after confirmation")
}

func TestHandleCleanup_WithAbort_DoesNotDeleteBranch(t *testing.T) {
    // Arrange
    // Mock stdin to simulate user typing "N"
    oldStdin := os.Stdin
    r, w, err := os.Pipe()
    require.NoError(t, err)
    os.Stdin = r
    _, err = w.Write([]byte("N\n"))
    require.NoError(t, err)
    w.Close()
    localPath, cleanup := test.SetupTestRepo(t)
    defer cleanup()
    defer func() { os.Stdin = oldStdin }()

    repo, err := git.PlainOpen(localPath)
    require.NoError(t, err)
    wt, err := repo.Worktree()
    require.NoError(t, err)

    oldWhen := time.Now().Add(-30 * 24 * time.Hour)
    oldBranchName := config.BranchPrefix + "old-to-keep"
    test.CreateAndPushStashBranch(t, repo, wt, localPath, oldBranchName, "old.txt", "old", oldWhen)
    require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")}))
    test.FetchAll(t, repo)

    // Act
    err = HandleCleanup()

    // Assert
    require.NoError(t, err, "HandleCleanup should not error on abort")
    remote, err := repo.Remote("origin")
    require.NoError(t, err)
    refs, err := remote.List(&git.ListOptions{})
    require.NoError(t, err)

    var branchFound bool
    for _, ref := range refs {
        if ref.Name().String() == "refs/heads/"+oldBranchName {
            branchFound = true
            break
        }
    }
    assert.True(t, branchFound, "Expected old branch to remain after aborting")
}
