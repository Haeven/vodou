package main

import (
	"flag"
	"os"
)

// var (
// 	hadError        bool
// 	hadRuntimeError bool

// 	r = newRunner(os.Stdout, os.Stderr)
// )

func main() {
	var filePath string

	// Declare command-line flag 'filepath', store in filePath variable
	flag.StringVar(&filePath, "filepath", "", "File path")
	flag.Parse()

	// If file path is defined, open file
	if filePath != "" {
		runFile(filePath)
		// ...else run command as prompt
	} else {
		runPrompt()
	}
}

func runFile(path string) {
	file, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}
	print(file)
	// r.run(string(file))
}

func runPrompt() {

}
