package editor

// Cursor is a moveable cursor with column and line coordinates indexed from 1.
// It uses offsets starting from 0 to represent the Cursor's position within
// a document too wide or long to fit terminal window.
type Cursor struct {
	col, line             uint
	colOffset, lineOffset uint
}

func newCursor() *Cursor {
	return &Cursor{
		col:  1,
		line: 1,
	}
}

// Col returns the 1-indexed column component of the cursor's position.
func (c *Cursor) Col() uint {
	return c.col
}

// Line returns the 1-indexed line component of the cursor's position.
func (c *Cursor) Line() uint {
	return c.line
}

// ColOffset returns the cursor's column offset.
func (c *Cursor) ColOffset() uint {
	return c.colOffset
}

// LineOffset returns the the cursor's line offset.
func (c *Cursor) LineOffset() uint {
	return c.lineOffset
}

// X returns the 1-indexed X-coordinate of the cursor relative to the screen.
func (c *Cursor) X() uint {
	return c.col - c.colOffset
}

// x returns the 1-indexed Y-coordinate of the cursor relative to the screen.
func (c *Cursor) Y() uint {
	return c.line - c.lineOffset
}

func (c *Cursor) left(prevLineLen uint) {
	if c.col > 1 {
		c.col--
		return
	}
	if c.line > 1 {
		c.line--
		c.end(prevLineLen)
	}
}

func (c *Cursor) home() {
	c.col = 1
}

func (c *Cursor) right(lineLen, nextLineLen, nLines uint) {
	if c.col <= lineLen {
		c.col++
		return
	}
	if c.line <= nLines {
		c.line++
		c.home()
	}
}

func (c *Cursor) end(lineLen uint) {
	c.col = lineLen + 1
}

// snap causes the cursor to snap to the end of the line if its current position
// would cause it to be rendered beyond the end of the line.
func (c *Cursor) snap(lineLen uint) {
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

func (c *Cursor) down(nLines uint) bool {
	if c.line > nLines {
		return false
	}
	c.line++
	return true
}

func (c *Cursor) pageUp(height uint) {
	c.line = c.lineOffset + 1
	for i := height; i > 0; i-- {
		if !c.up() {
			return
		}
	}
}

func (c *Cursor) pageDown(height, nLines uint) {
	nextLine := c.lineOffset + height
	if nextLine > nLines {
		c.line = nLines
	}
	for i := height; i > 0; i-- {
		if !c.down(nLines) {
			return
		}
	}
}

func (c *Cursor) scroll(width, height uint) {
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
