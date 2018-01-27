package main

import (
	"fmt"
	"log"
	"os"
	//
	"github.com/gsiems/go-isbn/pkg/isbn"
)

const (
	cShowHelp = iota
	cCheckDigit
	cParseValidate
)

func croak(msg string) {
	log.Fatal("ERROR: " + msg + "\n")
}

func carp(msg string) {
	log.Println("WARNING: " + msg)
}

func main() {

	action, inputs := parseArgs()

	if action == cShowHelp {
		showHelp()
	}

	if len(inputs) == 0 {
		croak("No ISBN supplied.")
	}

	if action == cCheckDigit {

		for _, val := range inputs {
			calcCheckDigit(val)
		}
	} else {

		xmlFile := os.Getenv("ISBN_RANGE_FILE")
		if xmlFile == "" {
			croak("ISBN_RANGE_FILE Env variable not set.")
		}

		_, err := isbn.LoadRangeData(xmlFile)
		if err != nil {
			croak(fmt.Sprintf("%s", err))
		}

		for _, val := range inputs {
			checkISBN(val)
		}
	}
}

func parseArgs() (action int, inputs []string) {

	args := os.Args[1:]
	for _, val := range args {
		if val == "-c" {
			if action == 0 {
				action = cCheckDigit
			}
		} else if val == "-p" {
			if action == 0 {
				action = cParseValidate
			}
		} else if val == "-h" {
			showHelp()
		} else {
			inputs = append(inputs, val)
		}
	}
	return action, inputs
}

func calcCheckDigit(input string) {

	testISBN := input
	if len(input) == 9 || len(input) == 12 {
		testISBN = input + "0"
	}

	result, err := isbn.CalcCheckDigit(testISBN)
	if err != nil {
		carp(fmt.Sprintf("%s", err))
		return
	}

	fmt.Printf("Check-digit for %s is %s\n", input, result)
}

func checkISBN(input string) {
	result, err := isbn.ParseISBN(input)
	if err != nil {
		carp(fmt.Sprintf("ISBN is invalid (%s)", err))
		return
	}
	fmt.Print("ISBN is valid: ")
	fmt.Println(result)
}

func showHelp() {

	fmt.Println(os.Args[0])
	fmt.Println("  Usage [-c|-p] isbn [isbn [isbn ...]]")
	fmt.Println()
	fmt.Println("    -h Show help")
	fmt.Println("    -c Calculate check-digit(s) (does not parse/validate)")
	fmt.Println("    -p Parse and validate ISBN(s)")
	os.Exit(0)
}
