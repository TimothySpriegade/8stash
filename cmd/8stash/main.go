package main

import (
	"fmt"
	"os"
	"strconv"

	flag "github.com/spf13/pflag"

	"8stash/internal/config"
	"8stash/internal/service"
	"8stash/internal/validation"
)

var operation string
var stashNumber int
var validationError error

func main() {
	os.Exit(Init())
}

func Init() int {

	operation, stashNumber, validationError = validation.ArgValidation(os.Args[1:])
	if validationError != nil {
		fmt.Fprintf(os.Stderr, "Argument error: %v\n", validationError)
		return 1
	}

	config.LoadConfig(config.ConfigName)

	switch operation {
	case "help":
		return help()
	case "push":
		return push()
	case "pop":
		return pop()
	case "list":
		return list()
	case "drop":
		return drop()
	case "cleanup":
		//using flagset here because i want to have specific flags for cleanup only
		cleanupCmd := flag.NewFlagSet("cleanup", flag.ExitOnError)
        var days int
		cleanupCmd.IntVarP(&days, "days", "d", config.CleanUpTimeInDays, "Override the cleanup retention period in days")
		cleanupCmd.Parse(os.Args[2:])
		return cleanup(days)
	default:
		fmt.Fprintf(os.Stderr, "Unknown operation: %v\n", operation)
		os.Exit(1)
	}
	return 0
}

func list() int {
	if err := service.HandleList(); err != nil {
		fmt.Println("Error fetching 8stashes")
		return 1
	}
	return 0
}

func help() int {
	service.PrintHelp()
	return 0
}

func push() int {
	var stashName, err = service.HandlePush()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during push operation: %v\n", err)
		return 1
	}
	fmt.Printf("Changes stashed to new branch: %s\n", stashName)
	return 0
}

func pop() int {
	if err := service.HandlePop(strconv.Itoa(stashNumber)); err != nil {
		fmt.Fprintf(os.Stderr, "Error during pop operation: %v\n", err)
		return 1
	}
	return 0
}

func drop() int {
	if err := service.HandleDrop(strconv.Itoa(stashNumber)); err != nil {
		return 1
	}
	return 0
}

func cleanup(days int) int {
	config.UpdateCleanupRetentionTime(days)
	if err := service.HandleCleanup(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during cleanup operation: %v\n", err)
		return 1
	}
	return 0
}
