package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var matches []byte
var matchesMap = map[int][]byte{
	0: nil,
}
var total int

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
	for k, v := range matchesMap {
		fmt.Println("key:", k, "value:", string(v))
	}
}

func matchLine(line []byte, pattern string) bool {

	if len(pattern) > 1 && pattern[0] == '^' {
		matched, consumed := matchHere(line, pattern[1:])
		return matched && consumed == len(pattern[1:])
	} else if len(pattern) > 1 && pattern[len(pattern)-1] == '$' {
		pat := pattern[:len(pattern)-1]
		matched, consumed := matchHere(line[len(line)-len(pat):], pat)
		return matched && consumed == len(pat)
	} else {
		for len(line) > 0 {
			clear(matchesMap)
			matches = nil
			if matched, consumed := matchHere(line, pattern); matched && consumed == len(pattern) {

				fmt.Println("full match was found with consumed length:", consumed)
				return true
			} else {
				fmt.Println("no full match. consumed:", consumed, "len(pattern):", len(pattern))
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

	if len(pattern) > 1 && pattern[0] == '(' {
		pat := pattern[1:strings.LastIndex(pattern, ")")]
		fmt.Println(pat)
		if strings.Contains(pat, "|") {
			alternatives := strings.Split(pat, "|")
			for _, alt := range alternatives {
				fmt.Println(alt)
				if subMatched, subConsumed := matchHere(line, alt); subMatched {
					remainingPattern := pattern[strings.LastIndex(pattern, ")")+1:]
					subMatchedAfter, subConsumedAfter := matchHere(line[subConsumed:], remainingPattern) // match pattern after ()
					consumedTotal := len(pattern) - len(remainingPattern)                                // add len of | and ()
					return subMatchedAfter, subConsumedAfter + consumedTotal
				}
			}
			return false, 0
		}
		if subMatched, subConsumed := matchHere(line, pat); subMatched {
			remainingPattern := pattern[strings.LastIndex(pattern, ")")+1:]

			maxKey := 0
			for k := range matchesMap {
				if k > maxKey {
					maxKey = k
				}
			}
			fmt.Println("appending to matchesMap")
			newKey := maxKey + 1
			matchesMap[newKey] = line[:subConsumed]
			subMatchedAfter, subConsumedAfter := matchHere(line[subConsumed:], remainingPattern)
			consumedTotal := len(pattern) - len(remainingPattern)
			return subMatchedAfter, subConsumedAfter + consumedTotal
		}
	}

	if len(pattern) > 2 && pattern[0] == '[' {
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
		seq := pattern[1]
		if len(pattern) > 2 && isQuantifier(pattern[2]) {
			i := quantifier(line, pattern, rune(pattern[2]))
			subMatched, subConsumed := matchHere(line[i:], pattern[3:])
			return subMatched, subConsumed + 3
		}
		if special(line, seq) {

			subMatched, subConsumed := matchHere(line[1:], pattern[2:])
			return subMatched, subConsumed + 2
		}
		if unicode.IsDigit(rune(seq)) {
			i, _ := strconv.Atoi(string(seq))
			ref := matchesMap[i]

			// if bytes.HasPrefix(line, ref) {
			// 	matches = append(matches, ref...)
			// 	subMatched, subConsumed := matchHere(line[len(ref):], pattern[2:])
			// 	fmt.Println("consumed:", subConsumed)
			// 	return subMatched, subConsumed + 2
			// }
			fmt.Println("it's digit number", i)
			fmt.Println("matching with", ref)
			if subMatched, subConsumed := matchHere(line, fmt.Sprintf("(%s)", ref)); subMatched {
				remainingPattern := pattern[2:]
				subConsumed -= 2
				subMatchedAfter, subConsumedAfter := matchHere(line[len(ref):], remainingPattern)
				return subMatchedAfter, 2 + subConsumedAfter // consume '\' and digit + everything afterwards
			}
		}
		return false, 0
	}

	if len(line) > 0 && len(pattern) > 1 && isQuantifier(pattern[1]) { // handle quantifiers
		i := quantifier(line, pattern, rune(pattern[1]))
		if pattern[1] == '+' && i == 0 {
			return false, 0
		}
		subMatched, subConsumed := matchHere(line[i:], pattern[2:])
		return subMatched, subConsumed + 2
	}

	if len(line) > 0 && (pattern[0] == '.' || pattern[0] == line[0]) {
		fmt.Println("direct match:", string(line[0]))
		matches = append(matches, line[0])
		subMatched, subConsumed := matchHere(line[1:], pattern[1:])
		return subMatched, subConsumed + 1
	}
	fmt.Println("no match")
	return false, 0
}

func isQuantifier(b byte) bool {
	return bytes.ContainsAny([]byte{b}, "+?")
}

func contains(line []byte, str string) (bool, []byte) {
	var foo []byte
	fmt.Println("searching for positive pattern", str, "in", string(line))
	for _, b := range line {
		for _, r := range str {
			if b == byte(r) {
				foo = append(foo, b)
			}
		}
	}
	return len(foo) > 0, foo
}

func doesntContain(line []byte, str string) (bool, []byte) {
	var foo []byte
	var match bool
	fmt.Println("searching for negative pattern", str, "in", string(line))
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

func quantifier(line []byte, pattern string, q rune) int {
	switch q {
	case '+':

		if pattern[0] == '\\' {
			seq := pattern[1]
			i := 0
			for i < len(line) && special(line[i:], seq) { // count how many times a line matches the pattern
				i++
			}
			return i
		} else if line[0] == pattern[0] {

			fmt.Println("direct match:", string(line[0]))
			matches = append(matches, line[0])

			i := 0
			for i < len(line) && line[i] == pattern[0] { // count how many times a line matches the pattern
				matches = append(matches, line[i])
				i++
			}
			return i
		}
	case '?':
		if line[0] == pattern[0] { // if there's one match instance, skip the matching line, and pattern char and ?
			fmt.Println("direct match:", string(line[0]))
			matches = append(matches, line[0])

			return 1
		}
	}
	return 0
}

func special(line []byte, seq byte) bool {
	switch seq {
	case 'd':
		if len(line) > 0 && unicode.IsDigit(rune(line[0])) {
			fmt.Println("digit match:", string(line[0]))
			matches = append(matches, line[0])
			return true
		}
	case 'w':
		if len(line) > 0 && (unicode.IsLetter(rune(line[0])) || unicode.IsDigit(rune(line[0])) || line[0] == '_') {
			fmt.Println("word match:", string(line[0]))
			matches = append(matches, line[0])
			return true
		}
	}

	return false
}
