package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	flag "github.com/spf13/pflag"

	"8stash/internal/config"
	"8stash/internal/service"
	"8stash/internal/validation"
)

type commandInput struct {
	operation   string
	stashNumber int
	args        []string
}

func main() {
	os.Exit(Init())
}

func Init() int {
	input, err := parseCommandInput(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Argument error: %v\n", err)
		return 1
	}

	config.LoadConfig(config.ConfigName)
	return execute(input)
}

func execute(input commandInput) int {
	switch input.operation {
	case "help":
		return help()
	case "push":
		commitMessage, err := parsePushFlags(input.args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Argument error: %v\n", err)
			return 1
		}
		return push(commitMessage)
	case "pop":
		return pop(input.stashNumber)
	case "list":
		return list()
	case "drop":
		return drop(input.stashNumber)
	case "cleanup":
		days, confirmation, err := parseCleanupFlags(input.args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Argument error: %v\n", err)
			return 1
		}
		config.UpdateSkipConfirmations(confirmation)
		return cleanup(days)
	default:
		fmt.Fprintf(os.Stderr, "Unknown operation: %v\n", input.operation)
		return 1
	}
}

func parseCommandInput(allArgs []string) (commandInput, error) {
	rawArgs := []string{}
	if len(allArgs) > 1 {
		rawArgs = allArgs[1:]
	}

	operation, stashNumber, err := validation.ArgValidation(rawArgs)
	if err != nil {
		return commandInput{}, err
	}

	return commandInput{
		operation:   operation,
		stashNumber: stashNumber,
		args:        rawArgs,
	}, nil
}

func parsePushFlags(args []string) (string, error) {
	pushCmd := flag.NewFlagSet("push", flag.ContinueOnError)
	pushCmd.SetOutput(io.Discard)
	var commitMessage string
	pushCmd.StringVarP(&commitMessage, "message", "m", "", "Add a descriptive message to a stash")

	if err := pushCmd.Parse(commandArgs(args)); err != nil {
		return "", err
	}

	return commitMessage, nil
}

func parseCleanupFlags(args []string) (int, bool, error) {
	cleanupCmd := flag.NewFlagSet("cleanup", flag.ContinueOnError)
	cleanupCmd.SetOutput(io.Discard)
	var days int
	var confirmation bool
	cleanupCmd.IntVarP(&days, "days", "d", config.CleanUpTimeInDays, "Override the cleanup retention period in days")
	cleanupCmd.BoolVarP(&confirmation, "yes", "y", config.SkipConfirmations, "Decide whether or not to skip the manual confirmation of stash deletion")

	if err := cleanupCmd.Parse(commandArgs(args)); err != nil {
		return 0, false, err
	}

	return days, confirmation, nil
}

func commandArgs(args []string) []string {
	if len(args) < 2 {
		return []string{}
	}
	return args[1:]
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

func push(commitMessage string) int {
	stashName, err := service.HandlePush(commitMessage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during push operation: %v\n", err)
		return 1
	}
	fmt.Printf("Changes stashed to new branch: %s\n", stashName)
	return 0
}

func pop(stashNumber int) int {
	if err := service.HandlePop(strconv.Itoa(stashNumber)); err != nil {
		fmt.Fprintf(os.Stderr, "Error during pop operation: %v\n", err)
		return 1
	}
	return 0
}

func drop(stashNumber int) int {
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
