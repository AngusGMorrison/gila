package bufio

import (
	"bytes"
	"fmt"
	"io"

	"github.com/angusgmorrison/gila/escseq"
	"github.com/angusgmorrison/gila/renderer"
)

const defaultBufferBytes = 4096

// TerminalWriter satisfies renderer.TerminalWriter. To minimize blocking IO, it
// writes to w only when flushed.
type TerminalWriter struct {
	buf *bytes.Buffer
	w   io.Writer
}

var _ renderer.TerminalWriter = (*TerminalWriter)(nil)

func NewTerminalWriter(w io.Writer) *TerminalWriter {
	return &TerminalWriter{
		buf: bytes.NewBuffer(make([]byte, 0, defaultBufferBytes)),
		w:   w,
	}
}

// Flush writes the contents of the TerminalWriter's buffer to its writer,
// returning any error that occurs.
func (tw *TerminalWriter) Flush() error {
	_, err := tw.buf.WriteTo(tw.w)
	return err
}

// Write appends p to the TerminalWriter's buffer, returning len(p) and a nil
// error.
func (tw *TerminalWriter) Write(p []byte) (int, error) {
	return tw.buf.Write(p)
}

// Write appends c to the TerminalWriter's buffer, returning a nil error.
func (tw *TerminalWriter) WriteByte(c byte) error {
	return tw.buf.WriteByte(c)
}

// Write appends r to the TerminalWriter's buffer, returning r's size in bytes and a nil error.
func (tw *TerminalWriter) WriteRune(r rune) (int, error) {
	return tw.buf.WriteRune(r)
}

// Write appends s to the TerminalWriter's buffer, returning len(s) and a nil
// error.
func (tw *TerminalWriter) WriteString(s string) (int, error) {
	return tw.buf.WriteString(s)
}

// WriteEscapeSequence formats the given EscSeq with args and writes it to the
// TerminalWriter's buffer, returning the number of bytes written and a nil
// error.
func (tw *TerminalWriter) WriteEscapeSequence(esc escseq.EscSeq, args ...any) (int, error) {
	n, err := fmt.Fprintf(tw.buf, string(esc), args...)
	if err != nil {
		return n, fmt.Errorf("write escape sequence %q: %w", esc, err)
	}
	return n, nil
}
