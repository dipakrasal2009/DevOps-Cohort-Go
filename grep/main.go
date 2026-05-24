// create the grep command in go that takes a pattern and a list of files as arguments and prints the lines that match the pattern. The command should also support the following options:
// -n: Show line numbers
// -B N: Print N lines before each match
// -A N: Print N lines after each match
// -c: Count matching lines instead of printing them
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
)

var (
	showLineNumber = flag.Bool("n", false, "Show line numbers")
	before         = flag.Int("B", 0, "Print N lines before each match")
	after          = flag.Int("A", 0, "Print N lines after each match")
	countMatches   = flag.Bool("c", false, "Count matching lines")
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: go run main.go [-n] <pattern> <file1> <file2> ... <fileN>")
		os.Exit(1)
	}

	pattern := args[0]
	files := args[1:]

	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("Invalid pattern: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		err := grepFile(file, re, *showLineNumber, *countMatches, *before, *after)
		if err != nil {
			fmt.Printf("Error processing file %s: %v\n", file, err)
		}
	}
}

func grepFile(filename string, re *regexp.Regexp, showLineNumber bool, countMatches bool, before int, after int) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	lineNumber := 0
	matchCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		lineNumber++

		if re.MatchString(line) {
			matchCount++
			if !countMatches {
				printMatch(lines, lineNumber-1, showLineNumber, before, after)
			}
		}
	}

	if countMatches {
		fmt.Printf("%s: %d matches\n", filename, matchCount)
	}

	return scanner.Err()
}

func printMatch(lines []string, index int, showLineNumber bool, before int, after int) {
	start := max(0, index-before)
	end := min(len(lines), index+after+1)

	for i := start; i < end; i++ {
		if showLineNumber {
			fmt.Printf("%d: %s\n", i+1, lines[i])
		} else {
			fmt.Println(lines[i])
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//commands to compile and run the program
// go build -o grep main.go
// ./grep -n -B 2 -A 2 "pattern" file1.txt file2.txt
