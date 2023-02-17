package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

// Scan at most one UTF-8 character at a time.
const scanMaxBytes = 4

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

	editor := newEditor(os.Stdin, os.Stdout)
	for editor.processKeypress() {
	}
	if err := editor.Err(); err != nil {
		return err
	}

	return nil
}

type editor struct {
	scanner *bufio.Scanner
	out     io.Writer
	err     error
}

func newEditor(r io.Reader, w io.Writer) *editor {
	return &editor{
		scanner: newScanner(r),
		out:     w,
	}
}

func (e *editor) Err() error {
	return e.err
}

func (e *editor) processKeypress() bool {
	rawKey, err := e.readKey()
	if err != nil {
		e.err = err
		return false
	}
	if rawKey == nil {
		return false
	}

	// Check for commands in the ASCII range before attempting to interpret Unicode.
	if len(rawKey) == 1 {
		switch key(rawKey[0]) {
		case ctrlChord(keyQuit): // quit
			return false
		}
	}

	return true
}

func (e *editor) readKey() ([]byte, error) {
	if ok := e.scanner.Scan(); !ok {
		if err := e.scanner.Err(); err != nil {
			return nil, fmt.Errorf("scan rune: %w", err)
		}
		return nil, nil
	}

	return e.scanner.Bytes(), nil
}

func newScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanBuf := make([]byte, scanMaxBytes)
	scanner.Buffer(scanBuf, scanMaxBytes)
	scanner.Split(bufio.ScanRunes)
	return scanner
}

const (
	// ctrlMask can be combined with any other ASCII character code, CHAR, to represent Ctrl-CHAR.
	// This is because the terminal handles Ctrl combinations by zeroing bits 5 and 6 of CHAR
	// (indexed from 0).
	ctrlMask = 0x1f
)

func ctrlChord(k key) key {
	return k & ctrlMask
}

type key byte

const (
	keyQuit key = 'q'
)
