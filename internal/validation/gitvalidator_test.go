package validation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func TestIsGitRepository_NotARepo(t *testing.T) {
	// Arrange
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmp, err := os.MkdirTemp("", "nogit-")
	if err != nil {
		t.Fatalf("mkdtemp failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(orig)
		_ = os.RemoveAll(tmp)
	}()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	// Act
	err = IsGitRepository()

	// Assert
	if err == nil {
		t.Fatalf("expected error for non-repo directory, got nil")
	}
	if !errors.Is(err, git.ErrRepositoryNotExists) {
		t.Fatalf("expected git.ErrRepositoryNotExists, got %v", err)
	}
}

func TestIsGitRepository_IsRepo(t *testing.T) {
	// Arrange
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmp, err := os.MkdirTemp("", "gitrepo-")
	if err != nil {
		t.Fatalf("mkdtemp failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(orig)
		_ = os.RemoveAll(tmp)
	}()
	if _, err := git.PlainInit(tmp, false); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	// Act
	err = IsGitRepository()

	// Assert
	if err != nil {
		t.Fatalf("expected nil for repo directory, got %v", err)
	}
}

func TestHasChanges_CleanRepo(t *testing.T) {
	// Arrange
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmp, err := os.MkdirTemp("", "cleanrepo-")
	if err != nil {
		t.Fatalf("mkdtemp failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(orig)
		_ = os.RemoveAll(tmp)
	}()

	repo, err := git.PlainInit(tmp, false)
	if err != nil {
		t.Fatalf("init repo failed: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree failed: %v", err)
	}
	f := filepath.Join(tmp, "initial.txt")
	if err := os.WriteFile(f, []byte("initial"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	if _, err := wt.Add("initial.txt"); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	_, err = wt.Commit("init", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
	})
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	// Act
	err = HasChanges()

	// Assert
	if err == nil {
		t.Fatalf("expected error for clean repo, got nil")
	}
	if !strings.Contains(err.Error(), "no changes detected") {
		t.Fatalf("expected 'no changes detected' error, got %v", err)
	}
}

func TestHasChanges_DirtyRepo(t *testing.T) {
	// Arrange
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmp, err := os.MkdirTemp("", "dirtyrepo-")
	if err != nil {
		t.Fatalf("mkdtemp failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(orig)
		_ = os.RemoveAll(tmp)
	}()

	repo, err := git.PlainInit(tmp, false)
	if err != nil {
		t.Fatalf("init repo failed: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree failed: %v", err)
	}
	base := filepath.Join(tmp, "base.txt")
	if err := os.WriteFile(base, []byte("base"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	if _, err := wt.Add("base.txt"); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	_, err = wt.Commit("init", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
	})
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	// Make an untracked change
	if err := os.WriteFile(filepath.Join(tmp, "untracked.txt"), []byte("change"), 0o644); err != nil {
		t.Fatalf("write untracked failed: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	// Act
	err = HasChanges()

	// Assert
	if err != nil {
		t.Fatalf("expected nil for dirty repo, got %v", err)
	}
}
