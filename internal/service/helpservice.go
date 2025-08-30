package service

import "fmt"

const spacer = "--------------------------------------------------------------------------------------------------------"

func PrintHelp() {
	fmt.Println("Welcome to 8stash Help!")
	fmt.Println(spacer)
	fmt.Println("Available commands:")
	fmt.Printf("  %-20s %s\n", "help", "Show this help message")
	fmt.Printf("  %-20s %s\n", "pop <number?>", "Attempts to pop a remote stash if only one exists")
	fmt.Printf("  %-20s %s\n", "", "If there are multiple stashes, enter a stash hash to pop your desired stash")
	fmt.Printf("  %-20s %s\n", "push", "Pushes your current local changes to a new remote stash")
	fmt.Printf("  %-20s %s\n", "list", "List all current stashes with their respective numbers")
	fmt.Printf("  %-20s %s\n", "drop <number>", "Delete the remote stash with the specified number")
	fmt.Println(spacer)
	fmt.Println("default behavior if no command is provided:")
	fmt.Println("8stash will attempt to push your current local changes to a new remote stash.")
	fmt.Println(spacer)
}
