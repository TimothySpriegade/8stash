package service

import "fmt"

const spacer = "--------------------------------------------------------------------------------------------------------"

func PrintHelp() {
	fmt.Println("Welcome to 8stash Help!")
	fmt.Println(spacer)
	fmt.Println("Available commands:")
	fmt.Println("  help             	Show this help message")
	fmt.Println("  pop              	attempts to pop a remote stash if only one exists")
	fmt.Println("                   	if there are multiple stashes, it will list them and ask which one to pop")
	fmt.Println("  push             	pushes your current local changes to a new remote stash ")
	fmt.Println("  list             	List all current stashes with their respective numbers")
	fmt.Println("  delete <number>  	Delete the remote stash with the specified number")
	fmt.Println(spacer)
	fmt.Println("default behavior if no command is provided:")
	fmt.Println("8stash will attempt to push your current local changes to a new remote stash.")
	fmt.Println(spacer)
}
