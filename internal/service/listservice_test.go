package service

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"8stash/internal/config"
	"8stash/internal/test"
)

func TestHandleList_PrintsStashes(t *testing.T) {
	// Arrange
	localPath, cleanup := test.SetupTestRepo(t)
	defer cleanup()

	repo, err := git.PlainOpen(localPath)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	test.CreateAndPushStashBranch(t, repo, wt, localPath, config.BranchPrefix+"one", "one.txt", "1", time.Now().Add(-2*time.Hour))
	test.CreateAndPushStashBranch(t, repo, wt, localPath, config.BranchPrefix+"two", "two.txt", "2", time.Now().Add(-1*time.Hour))

	// ensure local has remote refs
	test.FetchAll(t, repo)

	// Act
	out := captureOutput(t, func() {
		err = HandleList()
	})

	// Assert
	require.NoError(t, err)
	assert.Contains(t, out, "Available stashes:")
	assert.Contains(t, out, config.BranchPrefix+"one")
	assert.Contains(t, out, config.BranchPrefix+"two")
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
