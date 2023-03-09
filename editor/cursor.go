package editor

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

func (c *Cursor) left(prevLineLen int) bool {
	if c.col > 1 {
		c.col--
		return true
	}
	if c.line > 1 {
		c.line--
		return c.end(prevLineLen)
	}
	return false
}

func (c *Cursor) home() bool {
	if c.col == 1 {
		return false
	}
	c.col = 1
	return true
}

func (c *Cursor) right(lineLen, nextLineLen, nLines int) bool {
	if c.col <= lineLen {
		c.col++
		return true
	}
	if c.line <= nLines {
		c.line++
		return c.home()
	}
	return false
}

func (c *Cursor) end(lineLen int) bool {
	if c.col == lineLen+1 {
		return false
	}
	c.col = lineLen + 1
	return true
}

// snap causes the cursor to snap to the end of the line if its current position
// would cause it to be rendered beyond the end of the line.
func (c *Cursor) snap(lineLen int) {
	if c.col > lineLen+1 {
		c.end(lineLen)

	}
}

func (c *Cursor) up() bool {
	if c.line <= 1 {
		return false
	}
	c.line--
	return true
}

func (c *Cursor) down(nLines int) bool {
	if c.line > nLines {
		return false
	}
	c.line++
	return true
}

func (c *Cursor) pageUp(height int) {
	c.line = c.lineOffset + 1
	for i := height; i > 0; i-- {
		if !c.up() {
			return
		}
	}
}

func (c *Cursor) pageDown(height, nLines int) {
	c.line = c.lineOffset + height - 1
	if c.line > nLines {
		c.line = nLines
	}
	for i := height; i > 0; i-- {
		if !c.down(nLines) {
			return
		}
	}
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
	if zeroIdxCol < c.colOffset+3 {
		var leftMargin int
		if zeroIdxCol >= defaultCursorMargin && c.colOffset >= defaultCursorMargin-1 {
			leftMargin = zeroIdxCol - defaultCursorMargin
		}
		c.colOffset = leftMargin
	}
	// Scroll right: if the cursor is right of the right margin, update the
	// offset so that it shows a full screen of text where the cursor is in the
	// rightmost column.
	if zeroIdxCol >= c.colOffset+width {
		c.colOffset = zeroIdxCol - width + 1
	}
}
