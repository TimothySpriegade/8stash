package test

import (
	"os"
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
