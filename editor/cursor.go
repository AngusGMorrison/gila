package editor

import "github.com/angusgmorrison/gila/intutil"

// defaultCursorMargin controls the number of characters between the left-hand
// edge of the screen and the cursor when scrolling left, allowing the user to
// view the characters immediately preceding the cursor.
const defaultCursorMargin = 3

// Cursor is a moveable cursor with column and line coordinates indexed from 1.
// It uses offsets starting from 0 to represent the Cursor's position within a
// document too wide or long to fit terminal window.
//
// Cursor's directional methods (e.g. left, right, home) return booleans
// indicating whether the requested movement was possible given the bounds of
// the editor. Requesting a movement that is not possible is not considered an
// error.
type Cursor struct {
	col, line             int
	colOffset, lineOffset int
}

func newCursor() *Cursor {
	return &Cursor{
		col:  1,
		line: 1,
	}
}

// Col returns the 1-indexed column component of the cursor's position.
func (c *Cursor) Col() int {
	return c.col
}

// Line returns the 1-indexed line component of the cursor's position.
func (c *Cursor) Line() int {
	return c.line
}

// ColOffset returns the cursor's column offset.
func (c *Cursor) ColOffset() int {
	return c.colOffset
}

// LineOffset returns the the cursor's line offset.
func (c *Cursor) LineOffset() int {
	return c.lineOffset
}

// X returns the 1-indexed X-coordinate of the cursor relative to the screen.
func (c *Cursor) X() int {
	return c.col - c.colOffset
}

// x returns the 1-indexed Y-coordinate of the cursor relative to the screen.
func (c *Cursor) Y() int {
	return c.line - c.lineOffset
}

func (c *Cursor) left(prevLineLen int) {
	if c.col > 1 {
		c.col--
		return
	}
	c.up()
	c.end(prevLineLen)
}

func (c *Cursor) right(lineLen, nLines int) {
	if c.col <= lineLen {
		c.col++
		return
	}
	c.down(nLines)
	c.home()
}

// home returns the cursor to the beginning of the line.
func (c *Cursor) home() {
	c.col = 1
}

// end moves the cursor to the end of the line.
func (c *Cursor) end(lineLen int) {
	c.col = lineLen + 1
}

// snap causes the cursor to snap to the end of the line if its current position
// would cause it to be rendered beyond the end of the line.
func (c *Cursor) snap(lineLen int) {
	if c.col > lineLen+1 {
		c.end(lineLen)
	}
}

func (c *Cursor) up() {
	if c.line <= 1 {
		return
	}
	c.line--
}

func (c *Cursor) down(nLines int) {
	if c.line > nLines {
		return
	}
	c.line++
}

func (c *Cursor) pageUp(height int) {
	// The target line is one full page above the top of the current page, plus
	// two to account for the line offset's zero index and to leave the top line
	// of the previous screen visible at the bottom of the new screen.
	targetLine := c.lineOffset - height + 2
	c.line = intutil.Max(1, targetLine)
}

func (c *Cursor) pageDown(height, nLines int) {
	// The target line is one full page below the bottom of the current page,
	// less one line to allow the last line of the previous screen to be visible
	// at the top of the new screen.
	targetLine := (c.lineOffset + height - 1) + height
	c.line = intutil.Min(nLines+1, targetLine)
}

func (c *Cursor) scroll(width, height int) {
	zeroIdxLine, zeroIdxCol := c.line-1, c.col-1
	// Scroll up: if the cursor is above the last-known offset, update the
	// offset to the current cursor position.
	if zeroIdxLine < c.lineOffset {
		c.lineOffset = zeroIdxLine
	}
	// Scroll down: if the cursor is below the height of the screen as measured
	// from the current line offset, update the offset so that it shows a full
	// screen of text where the cursor is in the final row.
	if zeroIdxLine >= c.lineOffset+height {
		c.lineOffset = zeroIdxLine - height + 1
	}
	// Scroll left: if the cursor is left of the left margin, update the offset
	// to the the current cursor position plus a margin that allows the user to
	// see a few characters preceding the cursor.
	if zeroIdxCol < c.colOffset+defaultCursorMargin {
		c.colOffset = intutil.Max(0, zeroIdxCol-defaultCursorMargin)
	}
	// Scroll right: if the cursor is right of the right margin, update the
	// offset so that it shows a full screen of text where the cursor is in the
	// rightmost column.
	if zeroIdxCol >= c.colOffset+width {
		c.colOffset = zeroIdxCol - width + 1
	}
}
