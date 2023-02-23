// Package renderer encapsulates types and functions required to render text
// editor output.
package renderer

import (
	"fmt"
	"io"
	"time"

	"github.com/angusgmorrison/gila/editor"
	"github.com/angusgmorrison/gila/escseq"
)

const statusMsgMaxDuration = 5 * time.Second

// TerminalWriter writes output to a terminal-like device.
type TerminalWriter interface {
	io.Writer

	Flush() error
	WriteByte(b byte) error
	WriteRune(r rune) (int, error)
	WriteString(s string) (int, error)
	WriteEscapeSequence(e escseq.EscSeq, args ...any) (int, error)
}

// Screen describes the screen to which output will be written.
type Screen struct {
	Width, Height uint
}

// Renderer satisfies editor.Renderer, formatting content and writing to its
// underlying TerminalWriter.
type Renderer struct {
	about  string
	w      TerminalWriter
	screen Screen
}

var _ editor.Renderer = (*Renderer)(nil)

func New(name, version string, tw TerminalWriter, screen Screen) *Renderer {
	screen.Height -= 2 // reserve two lines for status and message bars
	return &Renderer{
		about:  fmt.Sprintf("%s -- version %s", name, version),
		w:      tw,
		screen: screen,
	}
}

// Render a complete frame to the renderer's TerminalWriter.
func (r *Renderer) Render(frame editor.Frame) error {
	if _, err := r.w.WriteEscapeSequence(escseq.EscCursorHide); err != nil {
		return err
	}
	if _, err := r.w.WriteEscapeSequence(escseq.EscCursorTopLeft); err != nil {
		return err
	}
	if err := r.renderPage(frame.Cursor, frame.Lines); err != nil {
		return err
	}
	if err := r.renderStatusBar(frame.Filename, frame.Cursor.Line(), uint(len(frame.Lines))); err != nil {
		return err
	}
	if err := r.renderMessageBar(frame.StatusMsg, frame.LastStatusTime); err != nil {
		return err
	}
	if _, err := r.w.WriteEscapeSequence(escseq.EscCursorPosition, frame.Cursor.Y(), frame.Cursor.X()); err != nil {
		return err
	}
	if _, err := r.w.WriteEscapeSequence(escseq.EscCursorShow); err != nil {
		return err
	}
	return r.w.Flush()
}

// Clear wipes the terminal represented the renderer's TerminalWriter.
func (r *Renderer) Clear() error {
	if _, err := r.w.WriteEscapeSequence(escseq.EscScreenClear); err != nil {
		return err
	}
	if _, err := r.w.WriteEscapeSequence(escseq.EscCursorTopLeft); err != nil {
		return err
	}
	return r.w.Flush()
}

// renderPage renders a full page of text to w. If lines is empty, it renders the homepage.
func (r *Renderer) renderPage(cursor *editor.Cursor, lines []*editor.Line) error {
	if len(lines) == 0 {
		return r.renderHomepage()
	}
	return r.renderContent(cursor, lines)
}

// renderStatusBar renders a status bar in the second-last row of the screen. It
// renders the filename, current line number and total lines in inverted colors.
func (r *Renderer) renderStatusBar(filename string, line, totalLines uint) error {
	if _, err := r.w.WriteEscapeSequence(escseq.EscGRendInvertColors); err != nil {
		return err
	}

	lhs := fmt.Sprintf(" %.20s - %d lines", filename, totalLines)
	maxLHSLen := min(uint(len(lhs)), r.screen.Width-1) // leave room for at least one padding space on RHS
	if _, err := r.w.WriteString(lhs[:maxLHSLen]); err != nil {
		return err
	}

	rhs := fmt.Sprintf("%d/%d ", line, totalLines)
	for i := uint(maxLHSLen); i < r.screen.Width; {
		if r.screen.Width-i == uint(len(rhs)) {
			if _, err := r.w.WriteString(rhs); err != nil {
				return err
			}
			break
		} else {
			if err := r.w.WriteByte(' '); err != nil {
				return err
			}
			i++
		}
	}

	if _, err := r.w.WriteEscapeSequence(escseq.EscGRendRestore); err != nil {
		return err
	}
	return r.renderNewLine()
}

// renderMessageBar renders a status message bar in the last row of the screen,
// provided that the status message has not yet expired.
func (r *Renderer) renderMessageBar(msg string, lastStatusTime time.Time) error {
	maxLen := min(uint(len(msg)), r.screen.Width)
	if maxLen > 0 && time.Since(lastStatusTime) < statusMsgMaxDuration {
		if _, err := r.w.WriteString(msg[:maxLen]); err != nil {
			return err
		}
	}
	if _, err := r.w.WriteEscapeSequence(escseq.EscLineClearFromCursor); err != nil {
		return err
	}
	return nil
}

func (r *Renderer) renderHomepage() error {
	for y := uint(1); y <= r.screen.Height; y++ {
		if y == r.screen.Height/3 {
			if err := r.renderAbout(); err != nil {
				return err
			}
		} else {
			if err := r.renderEmptyLine(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Renderer) renderContent(cursor *editor.Cursor, lines []*editor.Line) error {
	for y := uint(1); y <= r.screen.Height; y++ {
		lineIdx := y + cursor.LineOffset() - 1
		// We leave an empty line at the bottom of the document for the user to
		// insert new content which is not represented in lines. Hence, we must
		// check the lineIdx against the number of "real" lines to avoid
		// OutOfBounds errors.
		if lineIdx < uint(len(lines)) {
			line := lines[lineIdx].String()
			if err := r.renderLine(cursor, line); err != nil {
				return err
			}
		} else {
			if err := r.renderEmptyLine(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Renderer) renderAbout() error {
	about := center(r.about, r.screen.Width)
	maxLen := min(uint(len(about)), r.screen.Width)
	if _, err := r.w.WriteString(about[:maxLen]); err != nil {
		return fmt.Errorf("render about message %q: %w", about[:maxLen], err)
	}
	return nil
}

func (r *Renderer) renderEmptyLine() error {
	if err := r.w.WriteByte('~'); err != nil {
		return fmt.Errorf("write bytes %q: %w", '~', err)
	}
	return r.renderNewLine()
}

func (r *Renderer) renderLine(cursor *editor.Cursor, line string) error {
	leftMargin := min(cursor.ColOffset(), uint(len(line)))
	line = line[leftMargin:]
	rightMargin := min(uint(len(line)), r.screen.Width)
	line = line[:rightMargin]
	if _, err := r.w.WriteString(line); err != nil {
		return fmt.Errorf("write %q: %w", line, err)
	}
	return r.renderNewLine()
}

// renderNewLine clears any text to the right of the cursor position remaining
// from a previous render, then inserts a carriage return and newline.
func (r *Renderer) renderNewLine() error {
	if _, err := r.w.WriteEscapeSequence(escseq.EscLineClearFromCursor); err != nil {
		return fmt.Errorf("clear line from cursor: %w", err)
	}
	if _, err := r.w.WriteString("\r\n"); err != nil {
		return fmt.Errorf("write CRLF: %w", err)
	}
	return nil
}

func center(s string, width uint) string {
	leftPadding := (int(width) + len(s)) / 2
	rightPadding := -int(width) // Go interprets negative values as padding from the right
	// Bring the right margin all the way over to the left, then add half
	// (screen width + string len) to push the text into the middle.
	return fmt.Sprintf("%*s", rightPadding, fmt.Sprintf("%*s", leftPadding, s))
}

func min(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}
