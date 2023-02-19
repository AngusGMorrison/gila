package bufio

import (
	"bufio"
	"fmt"
	"io"

	"github.com/angusgmorrison/gila/editor"
	"github.com/angusgmorrison/gila/escseq"
)

// TerminalWriter satisfies editor.TerminalWriter.
type TerminalWriter struct {
	*bufio.Writer
}

var _ editor.TerminalWriter = (*TerminalWriter)(nil)

func NewTerminalWriter(w io.Writer) *TerminalWriter {
	return &TerminalWriter{bufio.NewWriter(w)}
}

// WriteEscapeSequence formats the given EscSeq with args and writes it to the underlying writer,
// returning the number of bytes written and any error.
func (tw *TerminalWriter) WriteEscapeSequence(esc escseq.EscSeq, args ...any) (int, error) {
	n, err := fmt.Fprintf(tw, string(esc), args...)
	if err != nil {
		return n, fmt.Errorf("write escape sequence %q: %w", esc, err)
	}
	return n, nil
}
