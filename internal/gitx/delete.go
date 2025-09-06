package gitx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
)

func deleteLocal(branchName string, repo *git.Repository, localRefName plumbing.ReferenceName) error {
	fmt.Printf("trying to delete branch %s locally\n", branchName)

	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("could not read HEAD: %w", err)
	}
	if headRef.Name() == localRefName {
		return fmt.Errorf("cannot delete current branch %q. Please switch to another branch first", branchName)
	}

	err = repo.Storer.RemoveReference(localRefName)
	if err != nil && !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return fmt.Errorf("failed to delete local branch %q: %w", branchName, err)
	}
	if err == nil {
		fmt.Printf("Local branch '%s' was deleted successfully.", branchName)
	} else {
		fmt.Printf("Local branch '%s' not found or already deleted.", branchName)
	}
	return nil
}

func deleteRemote(branchName string, repo *git.Repository, remoteRefSpec config.RefSpec, remoteName string) error {
	fmt.Printf("trying to delete branch %s on remote\n", branchName)

	pushOptions := &git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   []config.RefSpec{remoteRefSpec},
	}

	fmt.Printf("Attempting to delete remote branch '%s' on '%s'...\n", branchName, remoteName)
	var err error = nil
	err = repo.Push(pushOptions)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to delete remote branch: %w\n", err)
	}

	fmt.Printf("Remote branch '%s' on '%s' deleted successfully or was not present.\n", branchName, remoteName)
	return nil
}

func DeleteBranch(branchName string) error {
	repo, _, _, remoteName, err := getRepoContext()
	if err != nil {
		return err
	}

	if strings.TrimSpace(branchName) == "" {
		return fmt.Errorf(branchNameMustNotEmptyErrorMsg)
	}

	// Delete the local branch
	localRefName := plumbing.NewBranchReferenceName(branchName)
	if err := deleteLocal(branchName, repo, localRefName); err != nil {
		return err
	}

	// Delete the remote branch
	remoteRefSpec := config.RefSpec(":" + localRefName.String())
	if err := deleteRemote(branchName, repo, remoteRefSpec, remoteName); err != nil {
		return err
	}

	return nil
}
