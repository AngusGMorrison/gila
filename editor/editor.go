package editor

import (
	"bufio"
	"fmt"
	"io"
)

// Scan at most one UTF-8 character at a time.
const scanMaxBytes = 4

type key byte

const (
	keyQuit key = 'q'
)

type escapeSequence string

const (
	escCursorHide          escapeSequence = "\x1b[?25l"
	escCursorShow          escapeSequence = "\x1b[?25h"
	escCursorTopLeft       escapeSequence = "\x1b[H"
	escScreenClear         escapeSequence = "\x1b[2J"
	escLineClearFromCursor escapeSequence = "\x1b[K"
)

// Config contains editor configuration data.
type Config struct {
	Name, Version string
	Width, Height uint
}

// Editor holds the state for a text editor. Its methods run the main loop for reading and writing
// input to and from a terminal.
type Editor struct {
	config   Config
	scanner  *bufio.Scanner
	out      *bufio.Writer
	readErr  error
	writeErr error
}

// New returns a new *Editor that reads from r and writes to w.
func New(r io.Reader, w io.Writer, config Config) *Editor {
	return &Editor{
		config:  config,
		scanner: newScanner(r),
		out:     bufio.NewWriter(w),
	}
}

// Run starts the editor loop. The editor will update the screen and process user input until
// commanded to quit or an error occurs.
func (e *Editor) Run() (err error) {
	defer e.clearScreen()

	for e.refreshScreen() && e.processKeypress() {
	}
	if e.readErr != nil {
		return e.readErr
	}
	if e.writeErr != nil {
		return e.writeErr
	}
	return nil
}

// processKeypress is designed to be called in a tight loop. By returning a boolean, it is easily
// incorporated into a loop condition. If an error occurs during the refresh, it is saved to
// (*editor).readErr, and processKeypress returns false.
func (e *Editor) processKeypress() bool {
	rawKey, err := e.readKey()
	if err != nil {
		e.readErr = err
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

func (e *Editor) readKey() ([]byte, error) {
	if ok := e.scanner.Scan(); !ok {
		if err := e.scanner.Err(); err != nil {
			return nil, fmt.Errorf("scan rune: %w", err)
		}
		return nil, nil
	}

	return e.scanner.Bytes(), nil
}

// refreshScreen is designed to be called in a tight loop. By returning a boolean, it is easily
// incorporated into a loop condition. If an error occurs during the refresh, it is saved to
// (*editor).writeErr, and refreshScreen returns false.
func (e *Editor) refreshScreen() bool {
	if err := e.writeEscapeSeq(escCursorHide); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.writeEscapeSeq(escCursorTopLeft); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.drawRows(); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.flush(); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.writeEscapeSeq(escCursorShow); err != nil {
		e.writeErr = err
		return false
	}
	return true
}

func (e *Editor) writeEscapeSeq(esc escapeSequence) error {
	if _, err := e.out.WriteString(string(esc)); err != nil {
		return fmt.Errorf("write escape sequence %q: %w", esc, err)
	}
	return nil
}

func (e *Editor) clearScreen() error {
	if err := e.writeEscapeSeq(escScreenClear); err != nil {
		return err
	}
	if err := e.writeEscapeSeq(escCursorTopLeft); err != nil {
		return err
	}
	return e.flush()
}

func (e *Editor) flush() error {
	if err := e.out.Flush(); err != nil {
		return fmt.Errorf("flush output buffer: %w", err)
	}
	return nil
}

func (e *Editor) drawRows() error {
	for y := uint(0); y < e.config.Height; y++ {
		if y == (e.config.Height / 3) {
			// Write the welcome message.
			welcome := e.welcomeMessage()
			centered := center(welcome, e.config.Width)
			// Truncate the welcome message if it is too long for the screen.
			truncated := centered[:min(len(centered), int(e.config.Width))]
			if _, err := e.out.WriteString(truncated); err != nil {
				return fmt.Errorf("write row: %w", err)
			}
		} else {
			if err := e.out.WriteByte('~'); err != nil {
				return fmt.Errorf("write row: %w", err)
			}
		}
		// Clear the remains of the old line that have not been overwritten.
		if err := e.writeEscapeSeq(escLineClearFromCursor); err != nil {
			return err
		}
		// Add a new line to all but the last line of the screen.
		if y < e.config.Height-1 {
			if _, err := e.out.WriteString("\r\n"); err != nil {
				return fmt.Errorf("write row: %w", err)
			}
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

func (e *Editor) welcomeMessage() string {
	return fmt.Sprintf("%s -- version %s", e.config.Name, e.config.Version)
}

func center(s string, width uint) string {
	leftPadding := (int(width) + len(s)) / 2
	rightPadding := -int(width) // Go interprets negative values as padding from the right
	// Bring the right margin all the way over to the left, then add half (screen width + string
	// len) to push the text into the middle.
	return fmt.Sprintf("%*s", rightPadding, fmt.Sprintf("%*s", leftPadding, s))
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
