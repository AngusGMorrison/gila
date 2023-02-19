package editor

import (
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/angusgmorrison/gila/escseq"
)

// KeyReader reads a single keystroke or chord from input and returns its raw bytes.
type KeyReader interface {
	ReadKey() ([]byte, error)
}

// TerminalWriter writes output to a terminal-like device.
type TerminalWriter interface {
	io.Writer

	Flush() error
	WriteByte(b byte) error
	WriteRune(r rune) (int, error)
	WriteString(s string) (int, error)
	WriteEscapeSequence(e escseq.EscSeq, args ...any) (int, error)
}

const (
	// ctrlMask can be combined with any other ASCII character code, CHAR, to represent Ctrl-CHAR.
	// This is because the terminal handles Ctrl combinations by zeroing bits 5 and 6 of CHAR
	// (indexed from 0).
	ctrlMask    = 0x1f
	keyEsc      = '\x1b'
	keyDown     = 'j'
	keyLeft     = 'h'
	keyPageUp   = 65365
	keyPageDown = 65366
	keyRight    = 'l'
	keyUp       = 'k'
	keyQuit     = 'q' & ctrlMask
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
	r              KeyReader
	w              TerminalWriter
	readErr        error
	writeErr       error
}

// New returns a new *Editor that reads from kr and writes to tw.
func New(kr KeyReader, tw TerminalWriter, config Config) *Editor {
	return &Editor{
		config:         config,
		r:              kr,
		w:              tw,
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
	rawKey, err := e.r.ReadKey()
	if err != nil {
		e.readErr = err
		return false
	}

	key := transliterateKeypress(rawKey)
	if key == 0 { // EOF, return without error
		return false
	}

	switch key {
	case keyQuit:
		return false
	case keyPageUp:
		for i := e.config.Height; i > 0; i-- {
			e.moveCursor(keyUp)
		}
	case keyPageDown:
		for i := e.config.Height; i > 0; i-- {
			e.moveCursor(keyDown)
		}
	case keyLeft, keyDown, keyUp, keyRight:
		e.moveCursor(key)
	}

	return true
}

// refreshScreen is designed to be called in a tight loop. By returning a boolean, it is easily
// incorporated into a loop condition. If an error occurs during the refresh, it is saved to
// (*editor).writeErr, and refreshScreen returns false.
func (e *Editor) refreshScreen() bool {
	if _, err := e.w.WriteEscapeSequence(escseq.EscCursorHide); err != nil {
		e.writeErr = err
		return false
	}
	if _, err := e.w.WriteEscapeSequence(escseq.EscCursorTopLeft); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.drawRows(); err != nil {
		e.writeErr = err
		return false
	}
	if _, err := e.w.WriteEscapeSequence(escseq.EscCursorPosition, e.cursorPosition.y, e.cursorPosition.x); err != nil {
		e.writeErr = err
		return false
	}
	if _, err := e.w.WriteEscapeSequence(escseq.EscCursorShow); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.w.Flush(); err != nil {
		e.writeErr = err
		return false
	}
	return true
}

func (e *Editor) clearScreen() error {
	if _, err := e.w.WriteEscapeSequence(escseq.EscScreenClear); err != nil {
		return err
	}
	if _, err := e.w.WriteEscapeSequence(escseq.EscCursorTopLeft); err != nil {
		return err
	}
	return e.w.Flush()
}

func (e *Editor) drawRows() error {
	for y := uint(0); y < e.config.Height; y++ {
		if y == (e.config.Height / 3) {
			// Write the welcome message.
			welcome := e.welcomeMessage()
			centered := center(welcome, e.config.Width)
			// Truncate the welcome message if it is too long for the screen.
			truncated := centered[:min(len(centered), int(e.config.Width))]
			if _, err := e.w.WriteString(truncated); err != nil {
				return fmt.Errorf("write row: %w", err)
			}
		} else {
			if err := e.w.WriteByte('~'); err != nil {
				return fmt.Errorf("write row: %w", err)
			}
		}
		// Clear the remains of the old line that have not been overwritten.
		if _, err := e.w.WriteEscapeSequence(escseq.EscLineClearFromCursor); err != nil {
			return err
		}
		// Add a new line to all but the last line of the screen.
		if y < e.config.Height-1 {
			if _, err := e.w.WriteString("\r\n"); err != nil {
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

// transliterateKeypress interprets a raw keypress or chord as a UTF-8-encoded rune.
func transliterateKeypress(kp []byte) rune {
	if len(kp) == 0 {
		return 0
	}

	// Transliterate escape sequences.
	if isEscapeSequence(kp) {
		switch len(kp) {
		case 4:
			if kp[3] == '~' {
				switch kp[2] {
				case '5':
					return keyPageUp
				case '6':
					return keyPageDown
				}
			}
		case 3:
			switch kp[2] {
			case 'A':
				return keyUp
			case 'B':
				return keyDown
			case 'C':
				return keyRight
			case 'D':
				return keyLeft
			}
		}
	}

	r, _ := utf8.DecodeRune(kp)
	return r
}

// isEscapeSequence returns true if the keypress represents an escape sequence. The escape key
// itself is not counted as an escape sequence, and isEscapeSequence will return false in this case.
func isEscapeSequence(keypress []byte) bool {
	if len(keypress) <= 1 {
		return false
	}
	if keypress[1] == '[' {
		return true
	}
	return false
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
