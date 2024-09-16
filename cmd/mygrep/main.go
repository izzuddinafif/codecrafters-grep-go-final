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
	matches = nil
	if len(pattern) > 1 && pattern[0] == '^' && pattern[len(pattern)-1] == '$' {
		// Strip both ^ and $ from the pattern
		pat := pattern[1 : len(pattern)-1]
		if matched, patternConsumed, lineConsumed := matchHere(line, pat); matched && patternConsumed == len(pat) && lineConsumed == len(line) {
			fmt.Println("full match was found with pattern consumed:", patternConsumed+2, "len(pattern):", len(pattern), "line consumed:", lineConsumed, "len(line):", len(line))
			return true
		} else {
			fmt.Println("no full match (with ^ and $). pattern consumed:", patternConsumed, "line consumed:", lineConsumed, "len(line):", len(line))
			fmt.Println("match(es):", string(matches))
			for k, v := range matchesMap {
				fmt.Println("key:", k, "value:", string(v))
			}
			return false
		}
	}
	if len(pattern) > 1 && pattern[0] == '^' {
		if matched, patternConsumed, lineConsumed := matchHere(line, pattern[1:]); matched && patternConsumed+1 == len(pattern) {

			fmt.Println("full match was found with pattern consumed:", patternConsumed+2, "len(pattern):", len(pattern)+1, "line consumed:", lineConsumed, "len(line):", len(line))
			return true
		} else {
			fmt.Println("no full match. pattern consumed:", patternConsumed+2, "len(pattern):", len(pattern), "line consumed:", lineConsumed, "len(line):", len(line))
			fmt.Println("match(es):", string(matches))
			for k, v := range matchesMap {
				fmt.Println("key:", k, "value:", string(v))
			}
		}
	} else if len(pattern) > 1 && pattern[len(pattern)-1] == '$' {
		pat := pattern[:len(pattern)-1]
		if matched, patternConsumed, lineConsumed := matchHere(line, pat); matched && patternConsumed+1 == len(pattern) && lineConsumed == len(line) {

			fmt.Println("full match was found with pattern consumed:", patternConsumed, "len(pattern):", len(pattern), "line consumed:", lineConsumed, "len(line):", len(line))
			return true
		} else {
			fmt.Println("no full match. pattern consumed:", patternConsumed, "len(pattern):", len(pattern), "line consumed:", lineConsumed, "len(line):", len(line))
			fmt.Println("match(es):", string(matches))
			for k, v := range matchesMap {
				fmt.Println("key:", k, "value:", string(v))
			}
		}
	} else {
		for len(line) > 0 {
			clear(matchesMap)
			matches = nil
			if matched, patternConsumed, lineConsumed := matchHere(line, pattern); matched && patternConsumed == len(pattern) {

				fmt.Println("full match was found with pattern consumed:", patternConsumed, "len(pattern):", len(pattern), "line consumed:", lineConsumed, "len(line):", len(line))
				return true
			} else {
				fmt.Println("no full match. pattern consumed:", patternConsumed, "len(pattern):", len(pattern), "line consumed:", lineConsumed, "len(line):", len(line))
				fmt.Println("match(es):", string(matches))
				for k, v := range matchesMap {
					fmt.Println("key:", k, "value:", string(v))
				}
			}
			line = line[1:]
		}
	}

	return false
}

func matchHere(line []byte, pattern string) (bool, int, int) {

	fmt.Println("matching", string(line), "vs", pattern)
	if len(pattern) == 0 {
		fmt.Println("empty pattern match")
		return true, 0, 0
	}

	if len(pattern) > 1 && pattern[0] == '(' {
		pat := pattern[1:strings.Index(pattern, ")")]
		fmt.Println(pat)
		if strings.Contains(pat, "|") {
			alternatives := strings.Split(pat, "|")
			for _, alt := range alternatives {
				fmt.Println(alt)
				if subMatched, subPatternConsumed, subLineConsumed := matchHere(line, alt); subMatched {
					maxKey := 0
					for k := range matchesMap {
						if k > maxKey {
							maxKey = k
						}
					}
					fmt.Println("appending to matchesMap", subLineConsumed)
					newKey := maxKey + 1
					matchesMap[newKey] = line[:subLineConsumed]
					remainingPattern := pattern[strings.LastIndex(pattern, ")")+1:]
					subMatchedAfter, subPatternConsumedAfter, subLineConsumedAfter := matchHere(line[subPatternConsumed:], remainingPattern) // match pattern after ()
					patternConsumedTotal := len(pattern) - len(remainingPattern)                                                             // add len of | and ()

					fmt.Println("pat cons:", subPatternConsumedAfter+patternConsumedTotal)
					return subMatchedAfter, subPatternConsumedAfter + patternConsumedTotal, subLineConsumedAfter + subLineConsumed
				}
			}
			return false, 0, 0
		}
		if subMatched, subPatternConsumed, subLineConsumed := matchHere(line, pat); subMatched {
			remainingPattern := pattern[strings.LastIndex(pattern, ")")+1:]

			maxKey := 0
			for k := range matchesMap {
				if k > maxKey {
					maxKey = k
				}
			}
			fmt.Println("appending to matchesMap", subLineConsumed)
			newKey := maxKey + 1
			matchesMap[newKey] = line[:subLineConsumed]
			fmt.Println(string(matchesMap[newKey]))
			subMatchedAfter, subPatternConsumedAfter, subLineConsumedAfter := matchHere(line[subLineConsumed:], remainingPattern)

			fmt.Println("pat cons in ():", 2+subPatternConsumed)
			return subMatchedAfter, subPatternConsumedAfter + 2 + subPatternConsumed, subLineConsumedAfter + subLineConsumed
		}
	}

	if len(pattern) > 2 && pattern[0] == '[' {
		end := strings.LastIndex(pattern, "]")
		if end == -1 {
			// Handle invalid patterns with no closing bracket
			return false, 0, 0
		}
		set := pattern[1:end]
		if len(pattern) > end+1 && isQuantifier(pattern[end+1]) {

			i := quantifier(line, pattern[:end+1], rune(pattern[end+1]))

			fmt.Println(i)
			if pattern[end+1] == '+' && i == 0 {
				return false, 0, 0
			}
			subMatched, subPatternConsumed, subLineConsumed := matchHere(line[i:], pattern[end+2:])
			fmt.Println("pat cons:", subPatternConsumed+end+2)
			return subMatched, subPatternConsumed + end + 2, subLineConsumed + i
		}
		if pattern[1] == '^' {
			set = pattern[2:end]
			if ok, b := doesntContain(line, set); ok {
				matches = append(matches, b...)
				fmt.Println("negative character group match found:", string(b))
				subMatched, subPatternConsumed, subLineConsumed := matchHere(line[1:], pattern[end+1:])
				return subMatched, subPatternConsumed + end + 1, subLineConsumed + 1
			}
		} else if ok, b := contains(line, set); ok {
			matches = append(matches, b...)
			fmt.Println("positive character group match found:", string(b))
			subMatched, subPatternConsumed, subLineConsumed := matchHere(line[1:], pattern[end+1:])
			return subMatched, subPatternConsumed + end + 1, subLineConsumed + 1
		}
		return false, 0, 0
	}

	if len(pattern) > 1 && pattern[0] == '\\' {
		seq := pattern[1]
		if len(pattern) > 2 && isQuantifier(pattern[2]) {
			i := quantifier(line, pattern, rune(pattern[2]))
			subMatched, subPatternConsumed, subLineConsumed := matchHere(line[i:], pattern[3:])

			fmt.Println("pat cons in \\.q:", 3+subPatternConsumed)
			return subMatched, subPatternConsumed + 3, subLineConsumed + i
		}
		if special(line, seq) {

			subMatched, subPatternConsumed, subLineConsumed := matchHere(line[1:], pattern[2:])
			fmt.Println("pat cons:", 2+subPatternConsumed)
			return subMatched, subPatternConsumed + 2, subLineConsumed + 1
		}
		if unicode.IsDigit(rune(seq)) {
			i, _ := strconv.Atoi(string(seq))
			ref := matchesMap[i]

			// if bytes.HasPrefix(line, ref) {
			// 	matches = append(matches, ref...)
			// 	subMatched, subPatternConsumed, lineConsumed := matchHere(line[len(ref):], pattern[2:])
			// 	fmt.Println("patternConsumed:", subPatternConsumed)
			// 	return subMatched, subPatternConsumed + 2
			// }
			fmt.Println("it's digit number", i)
			fmt.Println("matching with", string(ref))
			if subMatched, subPatternConsumed, subLineConsumed := matchHere(line, fmt.Sprintf("(%s)", ref)); subMatched {
				remainingPattern := pattern[2:]
				subPatternConsumed -= 2
				subMatchedAfter, subPatternConsumedAfter, subLineConsumedAfter := matchHere(line[subLineConsumed:], remainingPattern)
				fmt.Println("pat cons:", 2+subPatternConsumedAfter)
				return subMatchedAfter, 2 + subPatternConsumedAfter, subLineConsumedAfter + subLineConsumed // consume '\' and digit + everything afterwards
			}
		}
		return false, 0, 0
	}

	if len(line) > 0 && len(pattern) > 1 && isQuantifier(pattern[1]) { // handle quantifiers
		fmt.Println("heeyy")
		i := quantifier(line, pattern, rune(pattern[1]))
		if pattern[1] == '+' && i == 0 {
			return false, 0, 0
		}
		subMatched, subPatternConsumed, subLineConsumed := matchHere(line[i:], pattern[2:])
		return subMatched, subPatternConsumed + 2, subLineConsumed + i
	}

	if len(line) > 0 && (pattern[0] == '.' || pattern[0] == line[0]) {
		fmt.Println("direct match:", string(line[0]))
		matches = append(matches, line[0])
		subMatched, subPatternConsumed, subLineConsumed := matchHere(line[1:], pattern[1:])
		return subMatched, subPatternConsumed + 1, subLineConsumed + 1
	}
	fmt.Println("no match")
	return false, 0, 0
}

func isQuantifier(b byte) bool {
	return bytes.ContainsAny([]byte{b}, "+?")
}

func contains(line []byte, str string) (bool, []byte) {
	var foo []byte
	fmt.Println("searching for positive pattern", str, "in", string(line))
	for _, b := range line {
		var found bool
		for _, r := range str {
			if b == byte(r) {
				foo = append(foo, b)
				fmt.Println("found match", string(b), string(r))
				found = true
			}
		}
		if !found {
			break
		}
	}
	return len(foo) > 0, foo
}

func doesntContain(line []byte, str string) (bool, []byte) {
	var foo []byte
	fmt.Println("searching for negative pattern", str, "in", string(line))
	for _, b := range line {
		var match bool
		for _, r := range str {
			if b == byte(r) {
				match = true
			} else {
				fmt.Println("match not found", string(b), string(r))
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
		if pattern[0] == '[' {
			i := 0
			set := pattern[1:strings.LastIndex(pattern, "]")]
			fmt.Println(set, "is set")
			if pattern[1] == '^' {
				fmt.Println("here")
				if ok, b := doesntContain(line[i:], set[1:]); ok {
					matches = append(matches, b...)
					i = len(b)
					fmt.Println(i, "i", string(b))
				}
			} else {
				// CAUTION: this doesn't include all matches as a group capture (not capturing them and adding them to matchesMap), just returning number of matches.
				if ok, b := contains(line[i:], set); ok {
					matches = append(matches, b...)
					i = len(b)
					fmt.Println(i, "i", string(b))
				}
			}
			return i
		}
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
