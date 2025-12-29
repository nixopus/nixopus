package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/raghavyuva/octopack/packer"
)

func main() {
	var inputFile = flag.String("input", "", "Path to JSON input file (default: stdin)")
	var outputFile = flag.String("output", "", "Path to output Dockerfile file (default: stdout)")
	var validate = flag.Bool("validate", true, "Validate the JSON specification before generating")
	flag.Parse()

	// Read input
	var input io.Reader
	if *inputFile != "" {
		file, err := os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
	} else {
		input = os.Stdin
	}

	// Read all input
	data, err := io.ReadAll(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Unmarshal JSON
	var spec packer.DockerfileSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Validate if requested
	if *validate {
		result := packer.Validate(&spec)
		if result.HasErrors() {
			fmt.Fprintf(os.Stderr, "Validation errors:\n")
			for _, err := range result.Errors {
				fmt.Fprintf(os.Stderr, "  %s\n", err.Error())
			}
			os.Exit(1)
		}
		if len(result.Warnings) > 0 {
			for _, warn := range result.Warnings {
				fmt.Fprintf(os.Stderr, "Warning: %s\n", warn.Error())
			}
		}
	}

	// Generate Dockerfile
	dockerfile, err := packer.Generate(&spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Dockerfile: %v\n", err)
		os.Exit(1)
	}

	// Write output to file or stdout
	if *outputFile != "" {
		err := os.WriteFile(*outputFile, []byte(dockerfile+"\n"), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(dockerfile)
	}
}
