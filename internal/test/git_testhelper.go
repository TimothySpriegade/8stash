package test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/require"
)

func SetupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	remotePath, err := os.MkdirTemp("", "remote")
	require.NoError(t, err)
	_, err = git.PlainInit(remotePath, true)
	require.NoError(t, err)

	localPath, err := os.MkdirTemp("", "local")
	require.NoError(t, err)
	repo, err := git.PlainInit(localPath, false)
	require.NoError(t, err)

	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(localPath))

	require.NoError(t, repo.Storer.SetReference(
		plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName("main")),
	))

	wt, err := repo.Worktree()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("initial.txt", []byte("init"), 0o644))
	_, err = wt.Add("initial.txt")
	require.NoError(t, err)
	_, err = wt.Commit("initial", &git.CommitOptions{
		Author: &object.Signature{Name: "T", Email: "t@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	_, err = repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{remotePath}})
	require.NoError(t, err)
	require.NoError(t, repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/heads/main:refs/heads/main")},
	}))

	cleanup := func() {
		_ = os.Chdir(origWD)
		_ = os.RemoveAll(remotePath)
		_ = os.RemoveAll(localPath)
	}
	return localPath, cleanup
}

func CreateAndPushStashBranch(t *testing.T, repo *git.Repository, wt *git.Worktree, localPath, fullBranchName, fileName, content string, when time.Time) {
	t.Helper()

	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(fullBranchName),
		Create: true,
	}))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, fileName), []byte(content), 0o644))
	_, err := wt.Add(fileName)
	require.NoError(t, err)
	_, err = wt.Commit("stash "+fullBranchName, &git.CommitOptions{
		Author: &object.Signature{Name: "S", Email: "s@example.com", When: when},
	})
	require.NoError(t, err)
	require.NoError(t, repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/heads/" + fullBranchName + ":refs/heads/" + fullBranchName)},
	}))
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))
}

func FetchAll(t *testing.T, repo *git.Repository) {
	t.Helper()
	err := repo.Fetch(&git.FetchOptions{RemoteName: "origin"})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		require.NoError(t, err)
	}
}
