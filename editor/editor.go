// Package editor implements the core loop of a text editor with pluggable input
// and output sources.
package editor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/angusgmorrison/gila/escseq"
)

const (
	defaultFilename = "[Untitled]"
	// Preallocate memory to hold pointers to at least nLinesToPreallocate lines of
	// text.
	nLinesToPreallocate = 1024
)

// KeyReader reads a single keystroke or chord from input and returns its raw
// bytes.
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

// Logger represents the minimal set of methods used to log the editor's
// workings.
type Logger interface {
	Println(a ...any)
	Printf(fmt string, a ...any)
}

// keynum is an enumerable that incorporates all Unicode symbols and
// additionally defines representations for keys with special functions.
type keynum rune

const (
	keyDel keynum = iota + 1e6 // start the function key definitions beyond the Unicode range
	keyDown
	keyEnd
	keyHome
	keyLeft
	keyPageUp
	keyPageDown
	keyRight
	keyUp
)

// Chords.
const (
	// ctrlMask can be combined with any other ASCII character code, CHAR, to
	// represent Ctrl-CHAR. This is because the terminal handles Ctrl
	// combinations by zeroing bits 5 and 6 of CHAR (indexed from 0).
	ctrlMask  = 0x1f
	chordQuit = 'q' & ctrlMask
)

// Config contains editor configuration data.
type Config struct {
	Name, Version string
	Width, Height uint
}

// Editor holds the state for a text editor. Its methods run the main loop for
// reading and writing input to and from a terminal.
type Editor struct {
	config   Config
	cursor   *cursor
	filename string
	// The text in the buffer.
	lines    []*line
	r        KeyReader
	w        TerminalWriter
	readErr  error
	writeErr error
	logger   Logger // TODO: make logging debug-only
}

// New returns a new *Editor that reads from kr and writes to tw.
func New(kr KeyReader, tw TerminalWriter, config Config, logger Logger) *Editor {
	config.Height-- // reserve the last line of the screen for the status bar
	return &Editor{
		config:   config,
		r:        kr,
		w:        tw,
		filename: defaultFilename,
		cursor:   newCursor(),
		logger:   logger,
	}
}

// Run starts the editor loop. The editor will update the screen and process
// user input until commanded to quit or an error occurs.
func (e *Editor) Run(filepath string) (err error) {
	defer e.clearScreen()

	if filepath != "" {
		if err = e.open(filepath); err != nil {
			return err
		}
	}

	for e.render() && e.processKeypress() {
	}
	if e.readErr != nil {
		return e.readErr
	}
	if e.writeErr != nil {
		return e.writeErr
	}
	return nil
}

// open opens the file at path and reads its lines into memory.
func (e *Editor) open(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { err = f.Close() }()

	e.filename = filepath.Base(path)
	e.lines = make([]*line, 0, nLinesToPreallocate)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		e.lines = append(e.lines, newLine(scanner.Text(), e.logger))
	}
	if err = scanner.Err(); err != nil {
		return fmt.Errorf("scan line from %s: %w", path, err)
	}
	return nil // EOF
}

// processKeypress is designed to be called in a tight loop. By returning a
// boolean, it is easily incorporated into a loop condition. If an error occurs
// during the refresh, it is saved to (*editor).readErr, and processKeypress
// returns false.
func (e *Editor) processKeypress() bool {
	rawKey, err := e.r.ReadKey()
	if err != nil {
		e.readErr = err
		return false
	}
	e.logger.Printf("read raw key %q\n", string(rawKey))

	key := transliterateKeypress(rawKey)
	if key == 0 { // EOF, return without error
		return false
	}
	e.logger.Printf("transliterated %q to %q\n", string(rawKey), key)

	switch key {
	case chordQuit:
		return false
	case keyPageUp:
		for i := e.config.Height; i > 0; i-- {
			e.moveCursor(keyUp)
		}
	case keyPageDown:
		for i := e.config.Height; i > 0; i-- {
			e.moveCursor(keyDown)
		}
	case keyHome, keyEnd, keyLeft, keyDown, keyUp, keyRight:
		e.moveCursor(key)
	}

	return true
}

// render is designed to be called in a tight loop. By returning a
// boolean, it is easily incorporated into a loop condition. If an error occurs
// during the refresh, it is saved to (*editor).writeErr, and render
// returns false.
func (e *Editor) render() bool {
	e.cursor.scroll(e.config.Width, e.config.Height)
	if _, err := e.w.WriteEscapeSequence(escseq.EscCursorHide); err != nil {
		e.writeErr = err
		return false
	}
	if _, err := e.w.WriteEscapeSequence(escseq.EscCursorTopLeft); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.renderLines(); err != nil {
		e.writeErr = err
		return false
	}
	if err := e.renderStatusBar(); err != nil {
		e.writeErr = err
		return false
	}
	if _, err := e.w.WriteEscapeSequence(escseq.EscCursorPosition, e.cursor.y(), e.cursor.x()); err != nil {
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

func (e *Editor) renderLines() error {
	nLines := e.len()
	for y := uint(1); y <= e.config.Height; y++ {
		lineIdx := y + e.cursor.lineOffset - 1
		if lineIdx < nLines { // inside the text buffer
			currentLine := e.lines[lineIdx].render
			maxCol := min(int(e.cursor.colOffset), len(currentLine)) // TODO: Handle grapheme clusters
			scrolledLine := currentLine[maxCol:]
			renderableLen := min(len(scrolledLine), int(e.config.Width))
			if _, err := e.w.WriteString(string(scrolledLine[:renderableLen])); err != nil { // TODO: candidate for optimisation, avoiding string conversion
				return fmt.Errorf("write line: %w", err)
			}
		} else { // after the text buffer
			if len(e.lines) == 0 && y == (e.config.Height/3) { // display the welcome message
				welcome := center(e.welcomeMessage(), e.config.Width)
				// Truncate the welcome message if it is too long for the screen.
				writeMax := min(len(welcome), int(e.config.Width))
				if _, err := e.w.WriteString(welcome[:writeMax]); err != nil {
					return fmt.Errorf("write line: %w", err)
				}
			} else {
				if err := e.w.WriteByte('~'); err != nil {
					return fmt.Errorf("write line: %w", err)
				}
			}
		}

		// Clear the remains of the old line that have not been overwritten.
		if _, err := e.w.WriteEscapeSequence(escseq.EscLineClearFromCursor); err != nil {
			return err
		}
		if _, err := e.w.WriteString("\r\n"); err != nil {
			return fmt.Errorf("write line: %w", err)
		}
	}

	return nil
}

func (e *Editor) renderStatusBar() error {
	if _, err := e.w.WriteEscapeSequence(escseq.EscGRendInvertColors); err != nil {
		return err
	}
	lhs := fmt.Sprintf(" %.20s - %d lines", e.filename, e.len())
	maxLHSLen := min(len(lhs), int(e.config.Width)-1) // leave room for at least one padding space on RHS
	if _, err := e.w.WriteString(lhs[:maxLHSLen]); err != nil {
		return err
	}

	rhs := fmt.Sprintf("%d/%d ", e.cursor.line, e.len())
	for i := uint(maxLHSLen); i < e.config.Width; {
		if e.config.Width-i == uint(len(rhs)) {
			if _, err := e.w.WriteString(rhs); err != nil {
				return err
			}
			break
		} else {
			if err := e.w.WriteByte(' '); err != nil {
				return err
			}
			i++
		}
	}

	if _, err := e.w.WriteEscapeSequence(escseq.EscGRendRestore); err != nil {
		return err
	}
	return nil
}

func (e *Editor) moveCursor(key keynum) {
	curLineLen := e.currentLine().renderLen()
	switch key {
	case keyHome:
		e.cursor.home()
	case keyEnd:
		e.cursor.end(curLineLen)
	case keyLeft:
		e.cursor.left(e.prevLine().renderLen())
	case keyDown:
		e.cursor.down(e.len())
	case keyUp:
		e.cursor.up()
	case keyRight:
		e.cursor.right(curLineLen, e.nextLine().renderLen(), e.len())
	default:
		panic(fmt.Errorf("unrecognized cursor key %q", key))
	}

	e.cursor.snap(e.currentLine().renderLen())
}

func (e *Editor) currentLine() *line {
	if e.cursor.line > e.len() {
		return nil
	}
	return e.lines[e.cursor.line-1]
}

func (e *Editor) prevLine() *line {
	if e.cursor.line <= 1 {
		return nil
	}
	return e.lines[e.cursor.line-2]
}

func (e *Editor) nextLine() *line {
	if e.cursor.line >= e.len() {
		return nil
	}
	return e.lines[e.cursor.line]
}

func (e *Editor) len() uint {
	return uint(len(e.lines))
}

func (e *Editor) welcomeMessage() string {
	return fmt.Sprintf("%s -- version %s", e.config.Name, e.config.Version)
}

// transliterateKeypress interprets a raw keypress or chord as a UTF-8-encoded rune.
func transliterateKeypress(kp []byte) keynum {
	if len(kp) == 0 {
		return 0
	}
	// Transliterate escape sequences. Due to differences between terminal
	// emulators, there may be several ways to represent the same escape
	// sequence.
	if isEscapeSequence(kp) {
		if kp[1] == '[' {
			switch len(kp) {
			case 4:
				if kp[3] == '~' {
					switch kp[2] {
					case '1':
						return keyHome
					case '3':
						return keyDel
					case '4':
						return keyEnd
					case '5':
						return keyPageUp
					case '6':
						return keyPageDown
					case '7':
						return keyHome
					case '8':
						return keyEnd
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
				case 'H':
					return keyHome
				case 'F':
					return keyEnd
				}
			}
		} else if kp[1] == 'O' {
			switch kp[2] {
			case 'H':
				return keyHome
			case 'F':
				return keyEnd
			}
		}
	}

	r, _ := utf8.DecodeRune(kp)
	return keynum(r)
}

// isEscapeSequence returns true if the keypress represents an escape sequence.
// The escape key itself is not counted as an escape sequence, and
// isEscapeSequence will return false in this case.
func isEscapeSequence(keypress []byte) bool {
	if len(keypress) <= 1 {
		return false
	}
	if keypress[1] == '[' || keypress[1] == 'O' {
		return true
	}
	return false
}

func center(s string, width uint) string {
	leftPadding := (int(width) + len(s)) / 2
	rightPadding := -int(width) // Go interprets negative values as padding from the right
	// Bring the right margin all the way over to the left, then add half
	// (screen width + string len) to push the text into the middle.
	return fmt.Sprintf("%*s", rightPadding, fmt.Sprintf("%*s", leftPadding, s))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
