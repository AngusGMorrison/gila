// Package editor implements the core loop of a text editor with pluggable input
// and output sources.
package editor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/angusgmorrison/gila/intutil"
)

const (
	defaultFilename  = "[Untitled]"
	defaultStatusMsg = "Help: Ctrl-S = save | Ctrl-Q = quit"
	// Preallocate memory to hold pointers to at least nLinesToPreallocate lines of
	// text.
	nLinesToPreallocate = 1024
	// The user must quit twice in a row to quit an unsaved file.
	forceQuitThreshold = 2
)

// KeyReader reads a single keystroke or chord from input and returns its raw
// bytes.
type KeyReader interface {
	ReadKey() ([]byte, error)
}

// Frame contains all the data required to render a complete frame.
type Frame struct {
	Cursor         *Cursor
	Lines          []*Line
	Filename       string
	StatusMsg      string
	LastStatusTime time.Time
	Dirty          bool
}

// Renderer renders a frame to some arbitrary output.
type Renderer interface {
	Render(frame Frame) error
	Clear() error
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
	keyBackspace keynum = iota + 1e6 // start the function key definitions beyond the Unicode range
	keyLineFeed
	keyDel
	keyDown
	keyEnd
	keyEsc
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
	ctrlMask       = 0x1f
	chordBackspace = 'h' & ctrlMask
	chordRefresh   = 'l' & ctrlMask
	chordSave      = 's' & ctrlMask
	chordQuit      = 'q' & ctrlMask
)

// Config contains editor configuration data.
type Config struct {
	Width, Height int
}

// Editor holds the state for a text editor. Its methods run the main loop for
// reading and writing input to and from a terminal.
type Editor struct {
	config         Config
	cursor         *Cursor
	filepath       string
	filename       string
	promptBuf      *Line
	statusMsg      string
	lastStatusTime time.Time
	// The number of consecutive quit commands, used for force-quitting unsaved documents.
	quitCount int
	// The text in the buffer.
	lines    []*Line
	dirty    bool
	r        KeyReader
	renderer Renderer
	readErr  error
	writeErr error
	logger   Logger // TODO: make logging debug-only
}

// New returns a new *Editor that reads from kr and writes to tw.
func New(kr KeyReader, r Renderer, config Config, logger Logger) *Editor {
	config.Height -= 2 // reserve the last two lines of the screen for the status bar and status message
	return &Editor{
		config:         config,
		filename:       defaultFilename,
		r:              kr,
		renderer:       r,
		promptBuf:      newLine(),
		statusMsg:      defaultStatusMsg,
		lastStatusTime: time.Now(),
		cursor:         newCursor(),
		logger:         logger,
	}
}

// Run starts the editor loop. The editor will update the screen and process
// user input until commanded to quit or an error occurs.
func (e *Editor) Run(filepath string) (err error) {
	defer e.renderer.Clear() // TODO: Use a multierror to capture all possible errors.

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

	e.filepath = path
	e.filename = filepath.Base(path)
	e.lines = make([]*Line, 0, nLinesToPreallocate)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		e.lines = append(e.lines, newLineFromString(scanner.Text()))
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
	case chordSave:
		if !e.save() {
			return false
		}
	case chordQuit:
		e.quitCount++
		if e.canForceQuit() {
			return false
		}
		e.setStatus("WARNING: Unsaved changes. Ctrl-Q to force quit.")
		return true
	case keyHome, keyEnd, keyLeft, keyDown, keyUp, keyRight, keyPageUp, keyPageDown:
		e.moveCursor(key)
	case keyBackspace:
		e.backspace()
	case keyDel:
		e.delete()
	case keyLineFeed:
		e.newLine()
	case keyEsc, chordRefresh:
		// No-op.
	default:
		e.insertRune(rune(key))
	}

	// The consecutive quit count is reset each time a non-quit kepress occurs.
	e.quitCount = 0
	return true
}

func (e *Editor) canForceQuit() bool {
	return !e.dirty || e.quitCount >= forceQuitThreshold
}

// render is designed to be called in a tight loop. By returning a
// boolean, it is easily incorporated into a loop condition. If an error occurs
// during the render, it is saved to (*editor).writeErr, and render
// returns false.
func (e *Editor) render() bool {
	e.cursor.scroll(e.config.Width, e.config.Height)
	if err := e.renderer.Render(e.frame()); err != nil {
		e.writeErr = err
		return false
	}
	return true
}

func (e *Editor) prompt(msg string) bool {
	for {
		e.setStatus(msg, e.promptBuf.String())
		if !e.render() {
			return false
		}

		rawKey, err := e.r.ReadKey()
		if err != nil {
			e.readErr = err
			return false
		}

		key := transliterateKeypress(rawKey)
		if key == keyLineFeed {
			e.setStatus("")
			return true
		} else if key == keyEsc {
			e.setStatus("")
			e.promptBuf.clear()
			return true
		} else if key == keyBackspace || key == keyDel {
			e.promptBuf.deleteLastRune()
		} else if !unicode.IsControl(rune(key)) {
			e.promptBuf.appendRune(rune(key))
		}
	}
}

// frame returns the current frame.
func (e *Editor) frame() Frame {
	return Frame{
		Cursor:         e.cursor,
		Lines:          e.lines,
		Filename:       e.filename,
		StatusMsg:      e.statusMsg,
		LastStatusTime: e.lastStatusTime,
		Dirty:          e.dirty,
	}
}

func (e *Editor) moveCursor(key keynum) {
	curLineLen := e.currentLine().RuneLen()
	switch key {
	case keyPageUp:
		e.cursor.pageUp(e.config.Height)
	case keyPageDown:
		e.cursor.pageDown(e.config.Height, e.len())
	case keyHome:
		e.cursor.home()
	case keyEnd:
		e.cursor.end(curLineLen)
	case keyLeft:
		e.cursor.left(e.prevLine().RuneLen())
	case keyDown:
		e.cursor.down(e.len())
	case keyUp:
		e.cursor.up()
	case keyRight:
		e.cursor.right(curLineLen, e.len())
	default:
		panic(fmt.Errorf("unrecognized cursor key %q", key))
	}

	e.cursor.snap(e.currentLine().RuneLen())
}

func (e *Editor) currentLine() *Line {
	if e.cursor.line > e.len() {
		return nil
	}
	return e.lines[e.cursor.line-1]
}

func (e *Editor) prevLine() *Line {
	if e.cursor.line <= 1 {
		return nil
	}
	return e.lines[e.cursor.line-2]
}

func (e *Editor) len() int {
	return len(e.lines)
}

func (e *Editor) insertRune(r rune) {
	line := e.currentLine()
	if line == nil {
		line = newLine()
		e.lines = append(e.lines, line)

	}
	line.insertRuneAt(r, e.cursor.col-1)
	e.cursor.col++
	e.dirty = true
}

func (e *Editor) backspace() {
	line := e.currentLine()
	if line == nil {
		return
	}

	// Deletion at the start of a line causes the current line to be merged into
	// the previous one.
	if e.cursor.col == 1 {
		e.mergeCurrentLineWithPrevious()
		return
	}

	line.deleteRuneAt(e.cursor.col - 2)
	e.cursor.col--
	e.dirty = true
}

func (e *Editor) delete() {
	// Deletion at the end of of a line, causes the next line to be merged into
	// the current one.
	if e.cursor.col == e.currentLine().RuneLen()+1 {
		e.mergeNextLineWithCurrent()
		return
	}

	e.cursor.col++
	e.backspace()
	e.cursor.col--
}

func (e *Editor) mergeNextLineWithCurrent() {
	if e.cursor.line == len(e.lines) {
		return
	}
	e.cursor.line++
	e.mergeCurrentLineWithPrevious()
}

func (e *Editor) mergeCurrentLineWithPrevious() {
	line, prevLine := e.currentLine(), e.prevLine()
	if line == nil || prevLine == nil {
		return
	}

	prevLineLen := prevLine.RuneLen()
	prevLine.append(line)
	e.deleteCurrentLine()
	e.cursor.line--
	e.cursor.col = prevLineLen
}

func (e *Editor) deleteCurrentLine() {
	e.lines = append(e.lines[:e.cursor.line-1], e.lines[e.cursor.line:]...)
}

func (e *Editor) newLine() {
	if e.cursor.line > len(e.lines) {
		return
	}
	currentLine := e.currentLine()
	runesToCopy := currentLine.Runes()[e.cursor.col-1:]
	newLineCap := intutil.Max(len(runesToCopy), lineRunesToPreallocate)
	newLineRunes := make([]rune, len(runesToCopy), newLineCap)
	copy(newLineRunes, runesToCopy)
	currentLine.runes = currentLine.runes[:e.cursor.col-1]
	newLine := newLineFromRunes(newLineRunes)
	e.lines = append(e.lines[:e.cursor.line], append([]*Line{newLine}, e.lines[e.cursor.line:]...)...)
	e.cursor.line++
	e.cursor.col = 1
}

func (e *Editor) String() string {
	var builder strings.Builder
	for _, l := range e.lines {
		builder.WriteString(l.String())
		builder.WriteByte('\n')
	}
	return builder.String()
}

func (e *Editor) save() bool {
	if !e.dirty {
		return true
	}
	// If the document is new, prompt for a filename.
	if e.filename == defaultFilename {
		if !e.prompt("Save as: %s") { // IO error
			return false
		}
		if e.promptBuf.RuneLen() == 0 { // no filename provided
			e.setStatus("Save aborted")
			return true
		}
		e.filepath = e.promptBuf.String()
		e.filename = filepath.Base(e.filepath)
		e.promptBuf.clear()
	}

	f, err := os.OpenFile(e.filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		e.setStatus("Changes not saved! IO error: %s", err)
		return true
	}
	defer f.Close()

	document := e.String()
	if _, err := f.WriteString(document); err != nil {
		e.setStatus("Changes not saved! IO error: %s", err)
		return true
	}

	e.setStatus("Saved")
	e.dirty = false
	return true
}

func (e *Editor) setStatus(format string, a ...any) {
	e.statusMsg = fmt.Sprintf(format, a...)
	e.lastStatusTime = time.Now()
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

	// Map special characters to keys.
	switch kp[0] {
	case chordBackspace, 127:
		return keyBackspace
	case '\x04':
		return keyDel
	case '\x1b':
		return keyEsc
	case '\r':
		return keyLineFeed
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
