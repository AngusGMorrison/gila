package editor

// cursor is a moveable cursor with column and line coordinates indexed from 1.
// It uses offsets starting from 0 to represent the cursor's position within
// a document too wide or long to fit terminal window.
type cursor struct {
	col, line             uint
	colOffset, lineOffset uint
}

func newCursor() *cursor {
	return &cursor{
		col:  1,
		line: 1,
	}
}

// x returns the 1-indexed x-coordinate of the cursor relative to the screen.
func (c *cursor) x() uint {
	return c.col - c.colOffset
}

// x returns the 1-indexed y-coordinate of the cursor relative to the screen.
func (c *cursor) y() uint {
	return c.line - c.lineOffset
}

func (c *cursor) left(prevLineLen uint) {
	if c.col > 1 {
		c.col--
		return
	}
	if c.line > 1 {
		c.line--
		c.end(prevLineLen)
	}
}

func (c *cursor) home() {
	c.col = 1
}

func (c *cursor) right(lineLen uint) {
	if c.col <= lineLen {
		c.col++
	}
}

func (c *cursor) end(lineLen uint) {
	c.col = lineLen + 1
}

// snap causes the cursor to snap to the end of the line if its current position
// would cause it to be rendered beyond the end of the line.
func (c *cursor) snap(lineLen uint) {
	if c.col > lineLen+1 {
		c.end(lineLen)
	}
}

func (c *cursor) up() {
	if c.line > 1 {
		c.line--
	}
}

func (c *cursor) down(nLines uint) {
	if c.line <= nLines {
		c.line++
	}
}

func (c *cursor) scroll(width, height uint) {
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
	// to the the current cursor position.
	if zeroIdxCol < c.colOffset {
		c.colOffset = zeroIdxCol
	}
	// Scroll right: if the cursor is right of the right margin, update the
	// offset so that it shows a full screen of text where the cursor is in the
	// rightmost column.
	if zeroIdxCol >= c.colOffset+width {
		c.colOffset = zeroIdxCol - width + 1
	}
}
