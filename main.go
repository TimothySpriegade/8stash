package main

import (
	"fmt"
	"os"
	service "stashpass/service"
	validation "stashpass/validation"

	flag "github.com/spf13/pflag"
)

var operation string
var stashNumber int
var validationError error

func main() {
	flag.Parse()
	args := flag.Args()
	operation, stashNumber, validationError = validation.ArgValidation(args)
	if validationError != nil {
		panic(validationError)
	}

	switch operation {
	case "help":
		service.PrintHelp()
	case "push":
		fmt.Println("StashPass Push")
	case "pop":
		fmt.Println("StashPass Pop")
	case "list":
		fmt.Println("StashPass List")
	case "delete":
		fmt.Println("StashPass Delete")
	default:
		fmt.Println("Unknown operation \"%s\".\n", operation)
		os.Exit(1)
	}
}
