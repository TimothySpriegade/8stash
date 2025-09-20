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
	listOfStashes, listOfStashesWithAuthor, err := Retrieve8stashList()
	if err != nil {
		return err
	}
	printStashes(listOfStashes, listOfStashesWithAuthor)
	return nil
}

func Retrieve8stashList() (map[string]string,map[string]string, error) {
	if err := gitx.UpdateRepository(); err != nil {
		return nil,nil,err
	}
	mapOfListAndTime, mapOfListAndAuthor , err := gitx.GetBranchesWithStringName(config.BranchPrefix)
	if err != nil {
		return nil,nil,err
	}
	return mapOfListAndTime, mapOfListAndAuthor, nil
}

func printStashes(stashes map[string]string, stashesAndAuthors map[string]string) {
	if len(stashes) == 0 {
		fmt.Println("No stashes found.")
		return
	}
	fmt.Println("Available stashes:")
	fmt.Println("-------------------------------------------------------------------")

	var stashList []struct {
		name   string
		time   string
		author string
	}

	for name, time := range stashes {

		stashList = append(stashList, struct {
			name   string
			time   string
			author string
		}{name, time, stashesAndAuthors[name]})
	}

	sort.Slice(stashList, func(i, j int) bool {
		return stashList[i].name < stashList[j].name
	})

	for _, stash := range stashList {
		fmt.Printf("%-30s - %-15s | %s\n", stash.name, stash.time, stash.author)
	}
}
