package main

import (
	validation "StashPass/validation"
	"fmt"
	"os"

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
		fmt.Println("StashPass Help:")
	case "pop":
		fmt.Println("StashPass Pop")
	case "push":
		fmt.Println("StashPass Push")
	case "list":
		fmt.Println("StashPass List")
	case "delete":
		fmt.Println("StashPass Delete")
	default:
		fmt.Println("Unknown operation \"%s\".\n", operation)
		os.Exit(1)
	}
}
