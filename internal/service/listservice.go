package service

import (
	"8stash/internal/gitx"
	"fmt"
)

const branchPrefix = "8stash/"

func HandleList() error {
	if err := gitx.UpdateRepository(); err != nil {
		return err
	}

	listOfStashes, err := retrieve8stashList()
	if err != nil {
		return err
	}

	printStashes(listOfStashes)

	return listOfStashes
}

func retrieve8stashList() (string[], error) {
	listOfStashes, err := gitx.GetBranchesWithStringName(branchPrefix)

	return nil
}

func printStashes(listOfStashes string[]) {
	for _, stash := range listOfStashes {
		fmt.Println(stash)
	}
}
