package editor

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

// Read as many bytes as the length of the longest character or escape sequence we're prepared to
// handle. E.g. The key F5 is represented by the 5-byte escape sequence \x1b[15~.
const readMaxBytes = 5

type escapeSequence string

const (
	escCursorHide          escapeSequence = "\x1b[?25l"
	escCursorShow          escapeSequence = "\x1b[?25h"
	escCursorPosition      escapeSequence = "\x1b[%d;%dH"
	escCursorTopLeft       escapeSequence = "\x1b[H"
	escScreenClear         escapeSequence = "\x1b[2J"
	escLineClearFromCursor escapeSequence = "\x1b[K"
)

const (
	// ctrlMask can be combined with any other ASCII character code, CHAR, to represent Ctrl-CHAR.
	// This is because the terminal handles Ctrl combinations by zeroing bits 5 and 6 of CHAR
	// (indexed from 0).
	ctrlMask = 0x1f
	keyEsc   = '\x1b'
	keyDown  = 'j'
	keyLeft  = 'h'
	keyRight = 'l'
	keyUp    = 'k'
	keyQuit  = 'q' & ctrlMask
)

// position represents 1-indexed x- and y-coordinates on a terminal.
type position struct {
	x, y uint
}

// Config contains editor configuration data.
type Config struct {
	Name, Version string
	Width, Height uint
}

// Editor holds the state for a text editor. Its methods run the main loop for reading and writing
// input to and from a terminal.
type Editor struct {
	config         Config
	cursorPosition position
	reader         *bufio.Reader
	writer         *bufio.Writer
	// keyBuffer is a permanent slice of len readMaxBytes intended to minimize allocations when
	// reading multi-byte sequences from reader. Its contents are overwritten on each read.
	keyBuffer []byte
	readErr   error
	writeErr  error
}

// New returns a new *Editor that reads from r and writes to w.
func New(r io.Reader, w io.Writer, config Config) *Editor {
	return &Editor{
		config:         config,
		reader:         bufio.NewReaderSize(r, readMaxBytes),
		writer:         bufio.NewWriter(w),
		keyBuffer:      make([]byte, readMaxBytes),
		cursorPosition: position{1, 1},
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
	key, err := e.readKey()
	if err != nil {
		e.readErr = err
		return false
	}
	if key == 0 { // EOF, return without error
		return false
	}

	switch key {
	case keyQuit:
		return false
	case keyLeft, keyDown, keyUp, keyRight:
		e.moveCursor(key)
	}

	return true
}

func (e *Editor) readKey() (rune, error) {
	n, err := e.reader.Read(e.keyBuffer)
	if err != nil {
		return 0, err
	}

	if e.keyBuffer[0] == keyEsc {
		if n == 3 && e.keyBuffer[1] == '[' {
			switch e.keyBuffer[2] {
			case 'A':
				return keyUp, nil
			case 'B':
				return keyDown, nil
			case 'C':
				return keyRight, nil
			case 'D':
				return keyLeft, nil
			}
		}
		return keyEsc, nil
	}

	key, _ := utf8.DecodeRune(e.keyBuffer[:n])
	return key, nil
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
	if err := e.writeEscapeSeq(escCursorPosition, e.cursorPosition.y, e.cursorPosition.x); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.writeEscapeSeq(escCursorShow); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.flush(); err != nil {
		e.writeErr = err
		return false
	}
	return true
}

func (e *Editor) writeEscapeSeq(esc escapeSequence, args ...any) error {
	if _, err := fmt.Fprintf(e.writer, string(esc), args...); err != nil {
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
	if err := e.writer.Flush(); err != nil {
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
			if _, err := e.writer.WriteString(truncated); err != nil {
				return fmt.Errorf("write row: %w", err)
			}
		} else {
			if err := e.writer.WriteByte('~'); err != nil {
				return fmt.Errorf("write row: %w", err)
			}
		}
		// Clear the remains of the old line that have not been overwritten.
		if err := e.writeEscapeSeq(escLineClearFromCursor); err != nil {
			return err
		}
		// Add a new line to all but the last line of the screen.
		if y < e.config.Height-1 {
			if _, err := e.writer.WriteString("\r\n"); err != nil {
				return fmt.Errorf("write row: %w", err)
			}
		}
	}

	return nil
}

func (e *Editor) moveCursor(cursorKey rune) {
	switch cursorKey {
	case keyLeft:
		if e.cursorPosition.x > 1 {
			e.cursorPosition.x--
		}
	case keyDown:
		if e.cursorPosition.y < e.config.Height {
			e.cursorPosition.y++
		}
	case keyUp:
		if e.cursorPosition.y > 1 {
			e.cursorPosition.y--
		}
	case keyRight:
		if e.cursorPosition.x < e.config.Width {
			e.cursorPosition.x++
		}
	default:
		panic(fmt.Errorf("unrecognized cursor key %q", cursorKey))
	}
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
