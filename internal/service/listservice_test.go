package service

import (
    "bytes"
    "io"
    "os"
    "strings"
    "testing"
    "time"

    "github.com/go-git/go-git/v6"
    "github.com/go-git/go-git/v6/plumbing/object"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "8stash/internal/config"
    "8stash/internal/test"
)

func TestHandleList_PrintsStashesWithAuthorAndTime(t *testing.T) {
    // Arrange
    localPath, cleanup := test.SetupTestRepo(t)
    defer cleanup()

    repo, err := git.PlainOpen(localPath)
    require.NoError(t, err)
    wt, err := repo.Worktree()
    require.NoError(t, err)

    authorOne := &object.Signature{Name: "Author One", Email: "one@example.com", When: time.Now().Add(-2 * time.Hour)}
    authorTwo := &object.Signature{Name: "Author Two", Email: "two@example.com", When: time.Now().Add(-1 * time.Hour)}

    test.CreateAndPushStashBranchWithAuthor(t, repo, wt, localPath, config.BranchPrefix+"one", "one.txt", "1", authorOne)
    test.CreateAndPushStashBranchWithAuthor(t, repo, wt, localPath, config.BranchPrefix+"two", "two.txt", "2", authorTwo)

    test.FetchAll(t, repo)

    // Act
    var actErr error
    out := captureOutput(t, func() {
        actErr = HandleList()
    })

    // Assert
    require.NoError(t, actErr)
    assert.Contains(t, out, "Available stashes:")
    assert.Contains(t, out, "8stash/one                     - 2h ago          | Author One")
    assert.Contains(t, out, "8stash/two                     - 1h ago          | Author Two")
}

func TestHandleList_NoStashes(t *testing.T) {
    // Arrange
    _, cleanup := test.SetupTestRepo(t)
    defer cleanup()

    // Act
    var actErr error
    out := captureOutput(t, func() {
        actErr = HandleList()
    })

    // Assert
    require.NoError(t, actErr)
    assert.Equal(t, "No stashes found.\n", out)
}

func TestHandleList_SortsOutputByName(t *testing.T) {
    // Arrange
    localPath, cleanup := test.SetupTestRepo(t)
    defer cleanup()

    repo, err := git.PlainOpen(localPath)
    require.NoError(t, err)
    wt, err := repo.Worktree()
    require.NoError(t, err)

    author := &object.Signature{Name: "Test Author", Email: "test@example.com", When: time.Now()}
    test.CreateAndPushStashBranchWithAuthor(t, repo, wt, localPath, config.BranchPrefix+"zebra", "z.txt", "z", author)
    test.CreateAndPushStashBranchWithAuthor(t, repo, wt, localPath, config.BranchPrefix+"apple", "a.txt", "a", author)
    test.FetchAll(t, repo)

    // Act
    var actErr error
    out := captureOutput(t, func() {
        actErr = HandleList()
    })

    // Assert
    require.NoError(t, actErr)
    lines := strings.Split(strings.TrimSpace(out), "\n")
    require.GreaterOrEqual(t, len(lines), 4, "Expected at least 4 lines of output for header, separator, and two stashes")
    assert.Contains(t, lines[2], config.BranchPrefix+"apple", "First stash should be 'apple'")
    assert.Contains(t, lines[3], config.BranchPrefix+"zebra", "Second stash should be 'zebra'")
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