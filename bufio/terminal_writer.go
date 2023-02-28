package bufio

import (
	"bufio"
	"fmt"
	"io"

	"github.com/angusgmorrison/gila/escseq"
	"github.com/angusgmorrison/gila/renderer"
)

const defaultBufferBytes = 4096

// TerminalWriter satisfies renderer.TerminalWriter.
type TerminalWriter struct {
	w *bufio.Writer
}

var _ renderer.TerminalWriter = (*TerminalWriter)(nil)

func NewTerminalWriter(w io.Writer) *TerminalWriter {
	return &TerminalWriter{
		w: bufio.NewWriterSize(w, defaultBufferBytes),
	}
}

// Flush writes the contents of the TerminalWriter's buffer to its writer,
// returning any error that occurs.
func (tw *TerminalWriter) Flush() error {
	return tw.w.Flush()
}

// Write appends p to the TerminalWriter's buffer. If p is longer than the
// buffer, the buffer will be written and flushed to output as many times as
// required to fully consume p.
func (tw *TerminalWriter) Write(p []byte) (int, error) {
	return tw.w.Write(p)
}

// Write appends c to the TerminalWriter's buffer.
func (tw *TerminalWriter) WriteByte(c byte) error {
	return tw.w.WriteByte(c)
}

// Write appends r to the TerminalWriter's buffer. Triggers a flush if the rune
// is longer than the remaining bytes in the buffer.
func (tw *TerminalWriter) WriteRune(r rune) (int, error) {
	return tw.w.WriteRune(r)
}

// Write appends s to the TerminalWriter's buffer, returning len(s) and a nil
// error. If s is longer than the buffer, the buffer will be written and flushed
// to output as many times as required to fully consume s.
func (tw *TerminalWriter) WriteString(s string) (int, error) {
	return tw.w.WriteString(s)
}

// WriteEscapeSequence formats the given EscSeq with args and writes it to the
// TerminalWriter's buffer. If the formatted escape sequence is longer than the
// buffer, the buffer will be written and flushed to output as many times as
// required to fully consume the escape sequence.
func (tw *TerminalWriter) WriteEscapeSequence(esc escseq.EscSeq, args ...any) (int, error) {
	n, err := fmt.Fprintf(tw.w, string(esc), args...)
	if err != nil {
		return n, fmt.Errorf("write escape sequence %q: %w", esc, err)
	}
	return n, nil
}
