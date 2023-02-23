// Package escseq provides escape sequence constants for working with ANSI
// terminals.
package escseq

type EscSeq string

const (
	// Cursor
	EscCursorHide     EscSeq = "\x1b[?25l"
	EscCursorShow     EscSeq = "\x1b[?25h"
	EscCursorPosition EscSeq = "\x1b[%d;%dH"
	EscCursorTopLeft  EscSeq = "\x1b[H"
	// Graphic rendition
	EscGRendInvertColors EscSeq = "\x1b[7m"
	EscGRendRestore      EscSeq = "\x1b[m"
	// Line
	EscLineClearFromCursor EscSeq = "\x1b[K"
	// Screen
	EscScreenClear EscSeq = "\x1b[2J"
)

// MaxLenBytes is the length in bytes of the longest escape sequence we intend
// to handle. 8 bytes is longer than any kepress on a standard ~100-key QWERTY
// keyboard.
const MaxLenBytes = 8
