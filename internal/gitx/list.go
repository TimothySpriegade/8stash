package gitx

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v6/plumbing"
)

func GetBranchesWithStringName(prefix string) (map[string]string, error) {
	repo, _, _, _, err := getRepoContext()
	if err != nil {
		return nil, err
	}
	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}
	defer refs.Close()

	branches := make(map[string]string)
	now := time.Now()

	err = refs.ForEach(func(ref *plumbing.Reference) error {
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

			branches[branchName] = timeStr
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error processing references: %w", err)
	}

	return branches, nil
}
