package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// Scan at most one UTF-8 character at a time.
const scanMaxBytes = 4

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() error {
	scanner := newScanner(os.Stdin)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan rune: %w", err)
		}

		r, _ := utf8.DecodeRune(scanner.Bytes())
		if r == 'q' {
			break
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
