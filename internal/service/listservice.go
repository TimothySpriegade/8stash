package service

import (
	"fmt"
	"sort"

	"8stash/internal/config"
	"8stash/internal/gitx"
)

func HandleList() error {
	if err := gitx.UpdateRepository(); err != nil {
		return err
	}
	listOfStashes, listOfStashesWithAuthor, listOfStashesWithMessages, err := Retrieve8stashList()
	if err != nil {
		return err
	}
	printStashes(listOfStashes, listOfStashesWithAuthor, listOfStashesWithMessages)
	return nil
}

func Retrieve8stashList() (map[string]string, map[string]string, map[string]string, error) {
	if err := gitx.UpdateRepository(); err != nil {
		return nil, nil, nil, err
	}
	mapOfListAndTime, mapOfListAndAuthor, mapOfListAndMessage, err := gitx.GetBranchInformationMapsByPrefix(config.BranchPrefix)
	if err != nil {
		return nil, nil, nil, err
	}
	return mapOfListAndTime, mapOfListAndAuthor, mapOfListAndMessage, nil
}

func printStashes(stashes map[string]string, stashesAndAuthors map[string]string, stashesAndMessages map[string]string) {
	if len(stashes) == 0 {
		fmt.Println("No stashes found.")
		return
	}
	fmt.Println("Available stashes:")
	fmt.Println("-------------------------------------------------------------------")

	var stashList []struct {
		name    string
		time    string
		author  string
		message string
	}

	for name, time := range stashes {
		stashList = append(stashList, struct {
			name    string
			time    string
			author  string
			message string
		}{name, time, stashesAndAuthors[name], stashesAndMessages[name]})
	}

	sort.Slice(stashList, func(i, j int) bool {
		return stashList[i].name < stashList[j].name
	})

	for _, stash := range stashList {
		fmt.Printf("%-30s - %-15s - %-15s | %s\n", stash.name, stash.time, stash.author, stash.message)
	}
}
