package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"8stash/internal/config"
	"8stash/internal/test"
)

func TestInit_PushCommand_Succeeds(t *testing.T) {
	// Arrange
	restoreConfig := snapshotConfig(t)
	defer restoreConfig()

	localPath, cleanupRepo := test.SetupTestRepo(t)
	defer cleanupRepo()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	filePath := filepath.Join(localPath, "wip.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("work in progress"), 0o644))

	defer stubArgs(t, "8stash", "push")()

	// Act
	stdout, stderr, exitCode := runInit(t)

	// Assert
	require.Equal(t, 0, exitCode)
	assert.Empty(t, strings.TrimSpace(stderr))

	stashBranch := parseStashBranch(t, stdout)
	assert.True(t, strings.HasPrefix(stashBranch, config.BranchPrefix))

	refs := listRemoteRefs(t, repo)
	assert.True(t, refExists(refs, "refs/heads/"+stashBranch), "expected remote branch %s", stashBranch)

	head, err := repo.Head()
	require.NoError(t, err)
	assert.Equal(t, "refs/heads/main", head.Name().String())

	wt, err = repo.Worktree()
	require.NoError(t, err)
	status, err := wt.Status()
	require.NoError(t, err)
	assert.True(t, status.IsClean())
}

func TestInit_PopCommand_Succeeds(t *testing.T) {
	// Arrange
	restoreConfig := snapshotConfig(t)
	defer restoreConfig()

	localPath, cleanupRepo := test.SetupTestRepo(t)
	defer cleanupRepo()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	stashNumber := "123"
	fullBranch := config.BranchPrefix + stashNumber
	test.CreateAndPushStashBranch(t, repo, wt, localPath, fullBranch, "stash.txt", "stash contents", time.Now().Add(-2*time.Hour))
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")}))
	test.FetchAll(t, repo)

	defer stubArgs(t, "8stash", "pop", stashNumber)()

	// Act
	_, stderr, exitCode := runInit(t)

	// Assert
	require.Equal(t, 0, exitCode)
	assert.Empty(t, strings.TrimSpace(stderr))

	refs := listRemoteRefs(t, repo)
	assert.False(t, refExists(refs, "refs/heads/"+fullBranch), "stash branch should be removed after pop")

	data, err := os.ReadFile(filepath.Join(localPath, "stash.txt"))
	require.NoError(t, err)
	assert.Equal(t, "stash contents", string(data))
}

func TestInit_ListCommand_PrintsStashes(t *testing.T) {
	// Arrange
	restoreConfig := snapshotConfig(t)
	defer restoreConfig()

	localPath, cleanupRepo := test.SetupTestRepo(t)
	defer cleanupRepo()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	alpha := config.BranchPrefix + "alpha"
	zeta := config.BranchPrefix + "zeta"
	test.CreateAndPushStashBranch(t, repo, wt, localPath, alpha, "alpha.txt", "alpha", time.Now().Add(-3*time.Hour))
	test.CreateAndPushStashBranch(t, repo, wt, localPath, zeta, "zeta.txt", "zeta", time.Now().Add(-2*time.Hour))
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")}))
	test.FetchAll(t, repo)

	defer stubArgs(t, "8stash", "list")()

	// Act
	stdout, stderr, exitCode := runInit(t)

	// Assert
	require.Equal(t, 0, exitCode)
	assert.Empty(t, strings.TrimSpace(stderr))
	assert.Contains(t, stdout, "Available stashes")
	assert.Contains(t, stdout, alpha)
	assert.Contains(t, stdout, zeta)
}

func TestInit_DropCommand_RemovesBranch(t *testing.T) {
	// Arrange
	restoreConfig := snapshotConfig(t)
	defer restoreConfig()

	localPath, cleanupRepo := test.SetupTestRepo(t)
	defer cleanupRepo()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	stashNumber := "321"
	fullBranch := config.BranchPrefix + stashNumber
	test.CreateAndPushStashBranch(t, repo, wt, localPath, fullBranch, "drop.txt", "drop", time.Now().Add(-time.Hour))
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")}))
	test.FetchAll(t, repo)

	defer stubArgs(t, "8stash", "drop", stashNumber)()

	// Act
	_, stderr, exitCode := runInit(t)

	// Assert
	require.Equal(t, 0, exitCode)
	assert.Empty(t, strings.TrimSpace(stderr))

	refs := listRemoteRefs(t, repo)
	assert.False(t, refExists(refs, "refs/heads/"+fullBranch), "branch should be deleted after drop")
}

func TestInit_CleanupCommand_RemovesOldStashes(t *testing.T) {
	// Arrange
	restoreConfig := snapshotConfig(t)
	defer restoreConfig()

	localPath, cleanupRepo := test.SetupTestRepo(t)
	defer cleanupRepo()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	oldBranch := config.BranchPrefix + "old-clean"
	newBranch := config.BranchPrefix + "new-keep"
	test.CreateAndPushStashBranch(t, repo, wt, localPath, oldBranch, "old.txt", "old", time.Now().Add(-45*24*time.Hour))
	test.CreateAndPushStashBranch(t, repo, wt, localPath, newBranch, "new.txt", "new", time.Now().Add(-2*24*time.Hour))
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")}))
	test.FetchAll(t, repo)

	defer stubArgs(t, "8stash", "cleanup", "-d", "10", "-y")()

	// Act
	stdout, stderr, exitCode := runInit(t)

	// Assert
	require.Equal(t, 0, exitCode)
	assert.Empty(t, strings.TrimSpace(stderr))
	assert.Contains(t, stdout, "Cleanup completed successfully.")

	refs := listRemoteRefs(t, repo)
	assert.False(t, refExists(refs, "refs/heads/"+oldBranch), "old stash should be deleted")
	assert.True(t, refExists(refs, "refs/heads/"+newBranch), "new stash should remain")
}

func TestInit_HelpCommand_PrintsUsage(t *testing.T) {
	// Arrange
	restoreConfig := snapshotConfig(t)
	defer restoreConfig()

	_, cleanupRepo := test.SetupTestRepo(t)
	defer cleanupRepo()

	defer stubArgs(t, "8stash", "help")()

	// Act
	stdout, stderr, exitCode := runInit(t)

	// Assert
	require.Equal(t, 0, exitCode)
	assert.Empty(t, strings.TrimSpace(stderr))
	assert.Contains(t, stdout, "Usage:")
	assert.Contains(t, stdout, "Available Commands:")
}

func TestInit_PushCommand_WithMessage_StoresMessage(t *testing.T) {
	// Arrange
	restoreConfig := snapshotConfig(t)
	defer restoreConfig()

	localPath, cleanupRepo := test.SetupTestRepo(t)
	defer cleanupRepo()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)

	filePath := filepath.Join(localPath, "wip.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("work in progress"), 0o644))

	customMessage := "implementing new feature X"
	defer stubArgs(t, "8stash", "push", "-m", customMessage)()

	// Act
	stdout, stderr, exitCode := runInit(t)

	// Assert
	require.Equal(t, 0, exitCode)
	assert.Empty(t, strings.TrimSpace(stderr))

	stashBranch := parseStashBranch(t, stdout)
	assert.True(t, strings.HasPrefix(stashBranch, config.BranchPrefix))

	test.FetchAll(t, repo)
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(stashBranch), true)
	require.NoError(t, err)

	commit, err := repo.CommitObject(ref.Hash())
	require.NoError(t, err)
	assert.Equal(t, customMessage, commit.Message)
}

func runInit(t *testing.T) (string, string, int) {
	t.Helper()
	operation = ""
	stashNumber = 0
	validationError = nil
	return captureOutputs(t, func() int { return Init() })
}

func captureOutputs(t *testing.T, fn func() int) (string, string, int) {
	t.Helper()
	oldStdout, oldStderr := os.Stdout, os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	require.NoError(t, err)
	stderrR, stderrW, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = stdoutW
	os.Stderr = stderrW
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	exitCode := fn()

	require.NoError(t, stdoutW.Close())
	require.NoError(t, stderrW.Close())

	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, stdoutR)
	_, _ = io.Copy(&stderrBuf, stderrR)

	require.NoError(t, stdoutR.Close())
	require.NoError(t, stderrR.Close())

	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

func stubArgs(t *testing.T, args ...string) func() {
	t.Helper()
	orig := os.Args
	copied := make([]string, len(args))
	copy(copied, args)
	os.Args = copied
	return func() {
		os.Args = orig
	}
}

func snapshotConfig(t *testing.T) func() {
	t.Helper()
	origPrefix := config.BranchPrefix
	origRetention := config.CleanUpTimeInDays
	origSkip := config.SkipConfirmations
	origHashType := config.NamingHashType
	origHashRange := config.HashRange

	return func() {
		config.BranchPrefix = origPrefix
		config.CleanUpTimeInDays = origRetention
		config.SkipConfirmations = origSkip
		config.NamingHashType = origHashType
		config.HashRange = origHashRange
	}
}

func parseStashBranch(t *testing.T, stdout string) string {
	t.Helper()
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Changes stashed to new branch:") {
			branch := strings.TrimSpace(strings.TrimPrefix(line, "Changes stashed to new branch:"))
			if branch != "" {
				return branch
			}
		}
	}
	t.Fatalf("expected stash branch in output, got: %s", stdout)
	return ""
}

func listRemoteRefs(t *testing.T, repo *git.Repository) []*plumbing.Reference {
	t.Helper()
	remote, err := repo.Remote("origin")
	require.NoError(t, err)
	refs, err := remote.List(&git.ListOptions{})
	require.NoError(t, err)
	return refs
}

func refExists(refs []*plumbing.Reference, want string) bool {
	for _, ref := range refs {
		if ref.Name().String() == want {
			return true
		}
	}
	return false
}
