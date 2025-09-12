package service

import "fmt"

const spacer = "--------------------------------------------------------------------------------------------------------"
const formatString = "  %-20s %s\n"

func PrintHelp() {
	fmt.Println("Welcome to 8stash Help!")
	fmt.Println(spacer)
	fmt.Println("Available commands:")
	fmt.Printf(formatString, "help", "Show this help message")
	fmt.Printf(formatString, "pop <number?>", "Attempts to pop a remote stash if only one exists")
	fmt.Printf(formatString, "", "If there are multiple stashes, enter a stash hash to pop your desired stash")
	fmt.Printf(formatString, "push", "Pushes your current local changes to a new remote stash")
	fmt.Printf(formatString, "list", "List all current stashes with their respective numbers")
	fmt.Printf(formatString, "drop <number>", "Delete the remote stash with the specified number")
	fmt.Printf(formatString, "cleanup", "Deletes all stashes older than 30 days")
	fmt.Println(spacer)
	fmt.Println("default behavior if no command is provided:")
	fmt.Println("8stash will attempt to push your current local changes to a new remote stash.")
	fmt.Println(spacer)
}
