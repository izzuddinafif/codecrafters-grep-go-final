package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

var matches []byte

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok := matchLine(line, pattern)

	if !ok {
		fmt.Println("match not found")
		os.Exit(1)
	}
	fmt.Println("match found:", len(matches), "match(es)")
	fmt.Println("match(es):", string(matches))
}

func matchLine(line []byte, pattern string) bool {

	if len(pattern) > 1 && pattern[0] == '^' {
		if matched, consumed := matchHere(line, pattern[1:]); matched && consumed == len(pattern[1:]) {
			fmt.Println("full match was found")
			return true
		} else {
			return false
		}
	} else {
		for len(line) > 0 {
			if matched, consumed := matchHere(line, pattern); matched && consumed == len(pattern) {
				fmt.Println("full match was found")
				return true
			}
			line = line[1:]
		}
	}

	return false
}

func matchHere(line []byte, pattern string) (bool, int) {

	fmt.Println("matching", string(line), "vs", pattern)
	if len(pattern) == 0 {
		fmt.Println("empty pattern match")
		return true, 0
	}

	if len(pattern) > 1 && pattern[0] == '[' {
		end := strings.LastIndex(pattern, "]")
		if pattern[1] == '^' {
			if ok, b := doesntContain(line, pattern[2:end]); ok {
				matches = append(matches, b...)
				fmt.Println("negative character group match found:", string(b))
				subMatched, subConsumed := matchHere(line[1:], pattern[end+1:])
				return subMatched, subConsumed + end + 1
			}
		} else if ok, b := contains(line, pattern[1:end]); ok {
			matches = append(matches, b...)
			fmt.Println("positive character group match found:", string(b))
			subMatched, subConsumed := matchHere(line[1:], pattern[end+1:])
			return subMatched, subConsumed + end + 1
		}
		return false, 0
	}

	if len(pattern) > 1 && pattern[0] == '\\' {
		switch pattern[1] {
		case 'd':
			if len(line) > 0 && unicode.IsDigit(rune(line[0])) {
				fmt.Println("digit match :", string(line[0]))
				matches = append(matches, line[0])
				subMatched, subConsumed := matchHere(line[1:], pattern[2:])
				return subMatched, subConsumed + 2
			}
		case 'w':
			if len(line) > 0 && (unicode.IsLetter(rune(line[0])) || unicode.IsDigit(rune(line[0])) || line[0] == '_') {
				fmt.Println("word match :", string(line[0]))
				matches = append(matches, line[0])
				subMatched, subConsumed := matchHere(line[1:], pattern[2:])
				return subMatched, subConsumed + 2
			}
		}
		return false, 0
	}
	if len(line) > 0 && (pattern[0] == '.' || pattern[0] == line[0]) {
		fmt.Println("direct match = ", string(line[0]))
		matches = append(matches, line[0])
		subMatched, subConsumed := matchHere(line[1:], pattern[1:])
		return subMatched, subConsumed + 1
	}
	fmt.Println("no match")
	return false, 0
}

func contains(line []byte, str string) (bool, []byte) {
	var foo []byte
	var match bool
	fmt.Println("searching positive pattern", str, "in", string(line))
	for _, b := range line {
		for _, r := range str {
			if b == byte(r) {
				foo = append(foo, b)
				match = true
			}
		}
	}
	return match, foo
}

func doesntContain(line []byte, str string) (bool, []byte) {
	var foo []byte
	var match bool
	fmt.Println("searching negative pattern", str, "in", string(line))
	for _, b := range line {
		for _, r := range str {
			if b == byte(r) {
				match = true
			}
		}
		if !match {
			foo = append(foo, b)
		}
		match = false
	}
	fmt.Println(string(foo))
	return len(foo) > 0, foo
}
