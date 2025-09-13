package service

import (
	"fmt"
	"sort"

	"8stash/internal/constants"
	"8stash/internal/gitx"
)

func HandleList() error {
	if err := gitx.UpdateRepository(); err != nil {
		return err
	}
	listOfStashes, err := Retrieve8stashList()
	if err != nil {
		return err
	}
	printStashes(listOfStashes)
	return nil
}

func Retrieve8stashList() (map[string]string, error) {
	if err := gitx.UpdateRepository(); err != nil {
		return nil, err
	}
	mapOfListAndTime, err := gitx.GetBranchesWithStringName(constants.BranchPrefix)
	if err != nil {
		return nil, err
	}
	return mapOfListAndTime, nil
}

func printStashes(stashes map[string]string) {
	if len(stashes) == 0 {
		fmt.Println("No stashes found.")
		return
	}
	fmt.Println("Available stashes:")
	fmt.Println("------------------")

	var stashList []struct {
		name string
		time string
	}

	for name, time := range stashes {
		stashList = append(stashList, struct {
			name string
			time string
		}{name, time})
	}

	sort.Slice(stashList, func(i, j int) bool {
		return stashList[i].name < stashList[j].name
	})

	for _, stash := range stashList {
		fmt.Printf("%-30s - %s\n", stash.name, stash.time)
	}
}
