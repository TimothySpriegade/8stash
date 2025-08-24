package main

import (
	service "8stash/internal/service"
	"8stash/internal/validation"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

var operation string
var stashNumber int
var validationError error

func main() {
	os.Exit(Init())
}

func Init() int {
	flag.Parse()
	args := flag.Args()
	operation, stashNumber, validationError = validation.ArgValidation(args)
	if validationError != nil {
		panic(validationError)
	}

	switch operation {
	case "help":
		service.PrintHelp()
		return 0
	case "push":
		var stashName, err = service.HandlePush()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during push operation: %v\n", err)
			return 1
		}
		fmt.Printf("Changes stashed to new branch: %s\n", stashName)
		return 0
	case "pop":
		fmt.Println("8stash Pop")
	case "list":
		fmt.Println("8stash List")
	case "delete":
		fmt.Println("8stash Delete")
	default:
		fmt.Println("Unknown operation \"%s\".\n", operation)
		os.Exit(1)
	}
	return 0
}
