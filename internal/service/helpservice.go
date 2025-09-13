package service

import "fmt"

const spacer = "------------------------------------------------------------------------------------------------------------------------------"
const formatString = "  %-20s %s\n"

func PrintHelp() {
	fmt.Println("Welcome to 8stash Help!")
	fmt.Println(spacer)
	fmt.Println("Available commands:")
	fmt.Printf(formatString, "help", "Show this help message")
	fmt.Printf(formatString, "pop <number?>", "Attempts to pop a remote stash if only one exists")
	fmt.Printf(formatString, "", "If there are multiple stashes, enter a stash hash to pop your desired stash")
	fmt.Printf(formatString, "push", "Pushes your current local changes to a new remote stash")
	fmt.Printf(formatString, "", "This is also the default behavior if no command is entered")
	fmt.Printf(formatString, "list", "List all current stashes with their respective numbers")
	fmt.Printf(formatString, "drop <number>", "Delete the remote stash with the specified number")
	fmt.Printf(formatString, "cleanup", "Deletes all stashes older than 30 days except configured differently")
	fmt.Println(spacer)
	fmt.Println("If you want to customize your setup, you can add the following attributes to a .8stash.yaml file in your repository:")
	fmt.Printf(formatString, "retention_days", "This sets the number of days for the cleanup command")
	fmt.Printf(formatString, "branch_prefix", "Set your custom branch prefix instead of 8stash/")
	fmt.Println(spacer)
}
