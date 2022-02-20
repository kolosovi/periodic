package main

import (
	"flag"
	"fmt"
	"github.com/kolosovi/periodic/internal/generator"
	"os"
)

type args struct {
	name            string
	userType        string
	userTypePackage string
}

func parseArgs() (args, error) {
	parsedArgs := args{}
	flag.StringVar(
		&parsedArgs.userTypePackage,
		"type-package",
		"",
		"Package of the type being cached",
	)
	flag.Parse()
	nonFlagArgs := flag.Args()
	if len(nonFlagArgs) != 2 {
		return parsedArgs, fmt.Errorf(
			"expected exactly 2 non-option arguments, got %v",
			len(nonFlagArgs),
		)
	}
	parsedArgs.name = nonFlagArgs[0]
	parsedArgs.userType = nonFlagArgs[1]
	return parsedArgs, nil
}

func printParseErrAndUsage(err error) {
	fmt.Printf("cannot parse args: %v\n", err)
	fmt.Printf("usage: periodic cachename typename [--type-package=packagename]\n")
}

func main() {
	parsedArgs, err := parseArgs()
	if err != nil {
		printParseErrAndUsage(err)
		os.Exit(1)
	}
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("cannot get current working directory: %v", err)
		os.Exit(2)
	}
	err = generator.Generate(
		parsedArgs.name,
		parsedArgs.userType,
		parsedArgs.userTypePackage,
		workingDir,
	)
	if err != nil {
		fmt.Printf("could not generate cache: %v", err)
		os.Exit(2)
	}
}
