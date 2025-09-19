package service

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"8stash/internal/config"
	"8stash/internal/gitx"
)

func HandleCleanup() error {
	if err := gitx.UpdateRepository(); err != nil {
		return fmt.Errorf("updating repository: %w", err)
	}

	stashes, err := gitx.GetBranchesWithStringName(config.BranchPrefix)
	if err != nil {
		return fmt.Errorf("get branches with prefix %s: %w", config.BranchPrefix, err)
	}
	fmt.Println("Cleaning up old stashes...")

	if len(stashes) == 0 {
		fmt.Println("No stashes found to clean up.")
		return nil
	}

	filtered := filterBranches(stashes, config.CleanUpTimeInDays)
	if len(filtered) == 0 {
		fmt.Println("No stashes found older than the cleanup time.")
		return nil
	}

	fmt.Printf("Found %d stashes, checking for those older than %d days...\n", len(stashes), config.CleanUpTimeInDays)

	if !awaitConfirmation() {
		fmt.Printf("Aborting the cleanup of branches\n")
		return nil
	}

	for branch := range filtered {
		fmt.Printf("Dropping stash branch: %s\n", branch)
		if err := gitx.DeleteBranch(branch); err != nil {
			return fmt.Errorf("drop branch %s: %w", branch, err)
		}
	}

	fmt.Println("Cleanup completed successfully.")
	return nil
}

func filterBranches(branches map[string]string, ageLimit int) map[string]string {
	filtered := make(map[string]string)
	for branch, age := range branches {
		if ageInt, err := parseDayString(age); err != nil {
			continue
		} else if ageInt >= ageLimit {
			filtered[branch] = age
		}
	}
	return filtered
}

func parseDayString(dayString string) (int, error) {
	var days int
	_, err := fmt.Sscanf(dayString, "%d days ago", &days)
	if err != nil {
		return 0, fmt.Errorf("parsing day string %s: %w", dayString, err)
	}
	return days, nil
}

func awaitConfirmation() bool {
	if config.SkipConfirmations {
		return true
	}

	fmt.Printf("Would you like to continue? [Y/N]\n")
	buf := bufio.NewReader(os.Stdin)
	answer, err := buf.ReadBytes('\n')

	if err != nil {
		fmt.Printf("Something went wrong: %v\n", err)
		return false
	}

	trimmedAnswer := strings.TrimSpace(string(answer))

	if strings.ToLower(trimmedAnswer) == "y" {
		fmt.Printf("continue deleting\n")
		return true
	}

	return false
}
