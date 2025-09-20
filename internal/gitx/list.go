package gitx

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

const TimeAuthorSpacer = "~"

func GetBranchesWithStringName(prefix string) (map[string]string, map[string]string, error) {
	repo, _, _, _, err := getRepoContext()
	if err != nil {
		return nil,nil,  err
	}
	refs, err := repo.References()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get references: %w", err)
	}
	defer refs.Close()

	branchToTimeMap := make(map[string]string)
	branchToAuthorMap := make(map[string]string)
	now := time.Now()

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		return processReference(ref, repo, prefix, now, branchToTimeMap, branchToAuthorMap)
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error processing references: %w", err)
	}

	return branchToTimeMap, branchToAuthorMap ,nil
}

func processReference(ref *plumbing.Reference, repo *git.Repository, prefix string, now time.Time, brancheToTimeMap map[string]string, branchToAuthorMap map[string]string) error {
	if !ref.Name().IsRemote() || ref.Type() != plumbing.HashReference {
		return nil
	}
	short := ref.Name().Short()
	parts := strings.SplitN(short, "/", 2)
	if len(parts) != 2 {
		return nil
	}
	remoteName, branchName := parts[0], parts[1]
	if remoteName != "origin" {
		return nil
	}

	if prefix == "" || strings.HasPrefix(branchName, prefix) {
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return fmt.Errorf("failed to get commit for branch %s: %w", branchName, err)
		}
		timeSince := now.Sub(commit.Author.When)
		brancheToTimeMap[branchName] = calculateTimeString(timeSince)
		branchToAuthorMap[branchName] = commit.Author.Name

	}
	return nil
}

func calculateTimeString(timeSince time.Duration) string {
	var timeStr string
	days := int(timeSince.Hours() / 24)
	if days > 0 {
		timeStr = fmt.Sprintf("%d days ago", days)
	} else {
		hours := int(timeSince.Hours())
		if hours > 0 {
			timeStr = fmt.Sprintf("%dh ago", hours)
		} else {
			minutes := int(timeSince.Minutes())
			timeStr = fmt.Sprintf("%dmin ago", minutes)
		}
	}
	return timeStr
}
