package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"

	"golang.org/x/term"
)

// Scan at most one UTF-8 character at a time.
const scanMaxBytes = 4

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() (err error) {
	// Enable terminal raw mode to process each keypress as it happens.
	initialTermState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("enable terminal raw mode: %w", err)
	}
	defer func() { err = term.Restore(int(os.Stdin.Fd()), initialTermState) }()
	// In raw mode, the cursor won't return to the start of the next line after the terminal echoes
	// the command used to run the program, so we force the line feed.
	fmt.Print("\r")

	scanner := newScanner(os.Stdin)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan rune: %w", err)
		}

		r, _ := utf8.DecodeRune(scanner.Bytes())
		if r == 'q' {
			break
		}

		if unicode.IsControl(r) {
			fmt.Printf("%d\r\n", r)
		} else {
			fmt.Printf("%d (%q)\r\n", r, r)
		}
	}

	return nil
}

func newScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanBuf := make([]byte, scanMaxBytes)
	scanner.Buffer(scanBuf, scanMaxBytes)
	scanner.Split(bufio.ScanRunes)
	return scanner
}
