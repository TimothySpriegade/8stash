package validation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var operationStashArgsRequirement = map[string]bool{
	"pop":  false,
	"drop": true,
	"push": false,
	"list": false,
	"help": false,
}

func isValidOperation(op string) bool {
	normalizedString := strings.ToLower(op)
	_, ok := operationStashArgsRequirement[normalizedString]
	return ok
}

func stashNumberIsRequiered(op string) bool {
	return operationStashArgsRequirement[strings.ToLower(op)]
}

func ArgValidation(args []string) (string, int, error) {
	var operation string
	var stashNumber int

	if len(args) < 1 {
		fmt.Println("No operation provided attempting push")
		return "push", 0, nil
	}
	operation = args[0]
	if !isValidOperation(operation) {
		fmt.Println("Invalid operation: " + operation + ". If you need help, run 8stash help")
		return "", 0, errors.New("invalid operation")
	}

	hasStashNumberArg := len(args) > 1

	if stashNumberIsRequiered(operation) && !hasStashNumberArg {
		fmt.Printf("Error: The '%s' operation requires a stash number.\n", operation)
		return "", 0, errors.New("operation requires a stash number")
	}

	if len(args) > 1 {
		var err error
		stashNumber, err = strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Error: Invalid number provided for stash number.", err)
			return "", 0, err
		}
	}

	return strings.ToLower(operation), stashNumber, nil
}
