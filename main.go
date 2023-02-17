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

	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("get terminal size: %w", err)
	}
	config := editorConfig{width: uint(w), height: uint(h)}
	editor := newEditor(os.Stdin, os.Stdout, config)
	// Clear the editor screen on exit.
	defer func() { err = editor.clearScreen() }()

	for editor.processKeypress() {
	}
	if err := editor.Err(); err != nil {
		return err
	}

	return nil
}

type editorConfig struct {
	width, height uint
}

type editor struct {
	config  editorConfig
	scanner *bufio.Scanner
	out     *bufio.Writer
	err     error
}

func newEditor(r io.Reader, w io.Writer, config editorConfig) *editor {
	return &editor{
		config:  config,
		scanner: newScanner(r),
		out:     bufio.NewWriter(w),
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

func (e *editor) refreshScreen() error {
	if err := e.clearScreen(); err != nil {
		return err
	}
	if err := e.drawRows(); err != nil {
		return err
	}
	return nil
}

func (e *editor) clearScreen() error {
	if _, err := e.out.WriteString(string(escClearScreen)); err != nil {
		return fmt.Errorf("clear screen: %w", err)
	}
	if _, err := e.out.WriteString(string(escCursorTopLeft)); err != nil {
		return fmt.Errorf("position cursor: %w", err)
	}
	if err := e.out.Flush(); err != nil {
		return fmt.Errorf("flush screen clear: %w", err)
	}
	return nil
}

func (e *editor) drawRows() error {
	for y := uint(0); y < e.config.height; y++ {
		if _, err := e.out.WriteString("~\r\n"); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	if err := e.out.Flush(); err != nil {
		return fmt.Errorf("flush row: %w", err)
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

type escapeSequence string

const (
	escClearScreen   escapeSequence = "\x1b[2J"
	escCursorTopLeft escapeSequence = "\x1b[H"
)
