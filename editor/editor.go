// Package editor implements the core loop of a text editor with pluggable input
// and output sources.
package editor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"
)

const (
	defaultFilename  = "[Untitled]"
	defaultStatusMsg = "Help: Ctrl-Q = quit"
	// Preallocate memory to hold pointers to at least nLinesToPreallocate lines of
	// text.
	nLinesToPreallocate  = 1024
	statusMsgMaxDuration = 5 * time.Second
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
	Width, Height uint
}

// Editor holds the state for a text editor. Its methods run the main loop for
// reading and writing input to and from a terminal.
type Editor struct {
	config         Config
	cursor         *Cursor
	filename       string
	statusMsg      string
	lastStatusTime time.Time
	// The text in the buffer.
	lines    []*Line
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
		r:              kr,
		renderer:       r,
		filename:       defaultFilename,
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

	e.filename = filepath.Base(path)
	e.lines = make([]*Line, 0, nLinesToPreallocate)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		e.lines = append(e.lines, newLine(scanner.Text()))
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

// frame returns the current frame.
func (e *Editor) frame() Frame {
	return Frame{
		Cursor:         e.cursor,
		Lines:          e.lines,
		Filename:       e.filename,
		StatusMsg:      e.statusMsg,
		LastStatusTime: e.lastStatusTime,
	}
}

func (e *Editor) moveCursor(key keynum) {
	curLineLen := e.currentLine().RenderLen()
	switch key {
	case keyHome:
		e.cursor.home()
	case keyEnd:
		e.cursor.end(curLineLen)
	case keyLeft:
		e.cursor.left(e.prevLine().RenderLen())
	case keyDown:
		e.cursor.down(e.len())
	case keyUp:
		e.cursor.up()
	case keyRight:
		e.cursor.right(curLineLen, e.nextLine().RenderLen(), e.len())
	default:
		panic(fmt.Errorf("unrecognized cursor key %q", key))
	}

	e.cursor.snap(e.currentLine().RenderLen())
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

func (e *Editor) nextLine() *Line {
	if e.cursor.line >= e.len() {
		return nil
	}
	return e.lines[e.cursor.line]
}

func (e *Editor) len() uint {
	return uint(len(e.lines))
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
