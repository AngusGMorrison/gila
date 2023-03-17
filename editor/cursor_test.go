package editor

import (
	"reflect"
	"testing"
)

func Test_Cursor_Col(t *testing.T) {
	t.Parallel()

	c := newCursor()
	if c.Col() != 1 {
		t.Errorf("expected Col() to return 1, got %d", c.Col())
	}

	c.col++
	if c.Col() != 2 {
		t.Errorf("expected Col() to return 2, got %d", c.Col())
	}
}

func Test_Cursor_Line(t *testing.T) {
	t.Parallel()

	c := newCursor()
	if c.Line() != 1 {
		t.Errorf("expected Line() to return 1, got %d", c.Line())
	}

	c.line++
	if c.Line() != 2 {
		t.Errorf("expected Line() to return 2, got %d", c.Line())
	}
}

func Test_Cursor_ColOffset(t *testing.T) {
	t.Parallel()

	c := newCursor()
	if c.ColOffset() != 0 {
		t.Errorf("expected ColOffset() to return 0, got %d", c.ColOffset())
	}

	c.colOffset++
	if c.ColOffset() != 1 {
		t.Errorf("expected ColOffset() to return 1, got %d", c.ColOffset())
	}
}

func Test_Cursor_LineOffset(t *testing.T) {
	t.Parallel()
	c := newCursor()
	if c.LineOffset() != 0 {
		t.Errorf("expected LineOffset() to return 0, got %d", c.LineOffset())
	}

	c.lineOffset++
	if c.LineOffset() != 1 {
		t.Errorf("expected LineOffset() to return 1, got %d", c.LineOffset())
	}
}

func Test_Cursor_X(t *testing.T) {
	c := newCursor()
	if c.X() != 1 {
		t.Errorf("expected X() to return 1, got %d", c.X())
	}

	c.colOffset++
	if c.X() != 0 {
		t.Errorf("expected X() to return 0, got %d", c.X())
	}
}

func Test_Cursor_Y(t *testing.T) {
	c := newCursor()
	if c.Y() != 1 {
		t.Errorf("expected Y() to return 1, got %d", c.Y())
	}

	c.lineOffset++
	if c.Y() != 0 {
		t.Errorf("expected Y() to return 0, got %d", c.Y())
	}
}

func Test_Cursor_left(t *testing.T) {
	t.Run("when c.line == 1", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name       string
			c          *Cursor
			wantCursor *Cursor
		}{
			{
				name: "col > 1",
				c: &Cursor{
					col:  2,
					line: 1,
				},
				wantCursor: &Cursor{
					col:  1,
					line: 1,
				},
			},
			{
				name: "col == 1",
				c: &Cursor{
					col:  1,
					line: 1,
				},
				wantCursor: &Cursor{
					col:  1,
					line: 1,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				tc.c.left(0)
				if !reflect.DeepEqual(tc.c, tc.wantCursor) {
					t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
				}
			})
		}
	})

	t.Run("when c.line > 1", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name        string
			c           *Cursor
			prevLineLen int
			wantCursor  *Cursor
		}{
			{
				name: "col > 1",
				c: &Cursor{
					col:  2,
					line: 2,
				},
				prevLineLen: 1,
				wantCursor: &Cursor{
					col:  1,
					line: 2,
				},
			},
			{
				name: "col == 1",
				c: &Cursor{
					col:  1,
					line: 2,
				},
				prevLineLen: 1,
				wantCursor: &Cursor{
					col:  2,
					line: 1,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				tc.c.left(tc.prevLineLen)
				if !reflect.DeepEqual(tc.c, tc.wantCursor) {
					t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
				}
			})
		}
	})
}

func Test_Cursor_right(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		c          *Cursor
		lineLen    int
		nLines     int
		wantCursor *Cursor
	}{
		{
			name: "col < lineLen",
			c: &Cursor{
				col:  1,
				line: 1,
			},
			lineLen: 2,
			wantCursor: &Cursor{
				col:  2,
				line: 1,
			},
		},
		{
			name: "col == lineLen",
			c: &Cursor{
				col:  2,
				line: 1,
			},
			lineLen: 2,
			wantCursor: &Cursor{
				col:  3,
				line: 1,
			},
		},
		{
			name: "col > lineLen && c.line < nLines",
			c: &Cursor{
				col:  3,
				line: 1,
			},
			lineLen: 2,
			nLines:  2,
			wantCursor: &Cursor{
				col:  1,
				line: 2,
			},
		},
		{
			name: "col > lineLen && c.line == nLines",
			c: &Cursor{
				col:  3,
				line: 2,
			},
			lineLen: 2,
			nLines:  2,
			wantCursor: &Cursor{
				col:  1,
				line: 3,
			},
		},
		{
			name: "col > lineLen && c.line > nLines",
			c: &Cursor{
				col:  1,
				line: 3,
			},
			lineLen: 0,
			nLines:  2,
			wantCursor: &Cursor{
				col:  1,
				line: 3,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.c.right(tc.lineLen, tc.nLines)
			if !reflect.DeepEqual(tc.c, tc.wantCursor) {
				t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
			}
		})
	}
}

func Test_Cursor_home(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		c          *Cursor
		wantCursor *Cursor
	}{
		{
			name: "col == 1",
			c: &Cursor{
				col:  1,
				line: 1,
			},
			wantCursor: &Cursor{
				col:  1,
				line: 1,
			},
		},
		{
			name: "col > 1",
			c: &Cursor{
				col:  2,
				line: 1,
			},
			wantCursor: &Cursor{
				col:  1,
				line: 1,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.c.home()
			if !reflect.DeepEqual(tc.c, tc.wantCursor) {
				t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
			}
		})
	}
}

func Test_Cursor_end(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		c          *Cursor
		lineLen    int
		wantCursor *Cursor
	}{
		{
			name: "when the cursor is at the end of the line, it does not move the cursor",
			c: &Cursor{
				col:  2,
				line: 1,
			},
			lineLen: 1,
			wantCursor: &Cursor{
				col:  2,
				line: 1,
			},
		},
		{
			name: "when the cursor is not at the end of the line, it moves the cursor to the end",
			c: &Cursor{
				col:  1,
				line: 1,
			},
			lineLen: 2,
			wantCursor: &Cursor{
				col:  3,
				line: 1,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.c.end(tc.lineLen)
			if !reflect.DeepEqual(tc.c, tc.wantCursor) {
				t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
			}
		})
	}
}

func Test_Cursor_up(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		c          *Cursor
		wantCursor *Cursor
	}{
		{
			name: "line == 1",
			c: &Cursor{
				col:  1,
				line: 1,
			},
			wantCursor: &Cursor{
				col:  1,
				line: 1,
			},
		},
		{
			name: "line > 1",
			c: &Cursor{
				col:  1,
				line: 2,
			},
			wantCursor: &Cursor{
				col:  1,
				line: 1,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.c.up()
			if !reflect.DeepEqual(tc.c, tc.wantCursor) {
				t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
			}
		})
	}
}

func Test_Cursor_down(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		c          *Cursor
		nLines     int
		wantCursor *Cursor
	}{
		{
			name: "line > nLines",
			c: &Cursor{
				col:  1,
				line: 2,
			},
			nLines: 1,
			wantCursor: &Cursor{
				col:  1,
				line: 2,
			},
		},
		{
			name: "line < nLines",
			c: &Cursor{
				col:  1,
				line: 1,
			},
			nLines: 2,
			wantCursor: &Cursor{
				col:  1,
				line: 2,
			},
		},
		{
			name: "line == nLines",
			c: &Cursor{
				col:  1,
				line: 2,
			},
			nLines: 2,
			wantCursor: &Cursor{
				col:  1,
				line: 3,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.c.down(tc.nLines)
			if !reflect.DeepEqual(tc.c, tc.wantCursor) {
				t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
			}
		})
	}
}

func Test_Cursor_snap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		c          *Cursor
		lineLen    int
		wantCursor *Cursor
	}{
		{
			name: "if the cursor is past the end of its line, it snaps to the end of the line",
			c: &Cursor{
				col:  4,
				line: 1,
			},
			lineLen: 2,
			wantCursor: &Cursor{
				col:  3,
				line: 1,
			},
		},
		{
			name: "if the cursor is at the end of its line, it does nothing",
			c: &Cursor{
				col:  2,
				line: 1,
			},
			lineLen: 1,
			wantCursor: &Cursor{
				col:  2,
				line: 1,
			},
		},
		{
			name: "if the cursor is before the end of its line, it does nothing",
			c: &Cursor{
				col:  1,
				line: 1,
			},
			lineLen: 2,
			wantCursor: &Cursor{
				col:  1,
				line: 1,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.c.snap(tc.lineLen)
			if !reflect.DeepEqual(tc.c, tc.wantCursor) {
				t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
			}
		})
	}
}

func Test_Cursor_pageUp(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		c          *Cursor
		height     int
		wantCursor *Cursor
	}{
		{
			name: "when there are more lines to traverse than the height of the screen, " +
				"it moves the cursor one line below the top of the current screen",
			c: &Cursor{
				col:        1,
				line:       20,
				lineOffset: 15,
			},
			height: 5,
			wantCursor: &Cursor{
				col:        1,
				line:       12,
				lineOffset: 15,
			},
		},
		{
			name: "when there are fewer lines to traverse than the height of the screen, " +
				"it moves the cursor to the top of the document",
			c: &Cursor{
				col:        1,
				line:       10,
				lineOffset: 0,
			},
			height: 20,
			wantCursor: &Cursor{
				col:        1,
				line:       1,
				lineOffset: 0,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.c.pageUp(tc.height)
			if !reflect.DeepEqual(tc.c, tc.wantCursor) {
				t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
			}
		})
	}
}

func Test_Cursor_pageDown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		c          *Cursor
		height     int
		nLines     int
		wantCursor *Cursor
	}{
		{
			name: "when there are enough lines below the cursor position to fill the whole screen, " +
				"it moves the cursor to the bottom of the current screen",
			c: &Cursor{
				col:        1,
				line:       10,
				lineOffset: 5,
			},
			height: 5,
			nLines: 20,
			wantCursor: &Cursor{
				col:        1,
				line:       14,
				lineOffset: 5,
			},
		},
		{
			name: "when there are too few lines below the cursor position to fill a whole screen" +
				"it moves the cursor to the bottom of the document",
			c: &Cursor{
				col:        1,
				line:       10,
				lineOffset: 0,
			},
			height: 20,
			nLines: 10,
			wantCursor: &Cursor{
				col:        1,
				line:       11,
				lineOffset: 0,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.c.pageDown(tc.height, tc.nLines)
			if !reflect.DeepEqual(tc.c, tc.wantCursor) {
				t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
			}
		})
	}
}

func Test_Cursor_scroll(t *testing.T) {
	t.Parallel()

	t.Run("scroll up", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name       string
			c          *Cursor
			wantCursor *Cursor
		}{
			{
				name: "when the cursor's zero-indexed line is above the current line offset, " +
					"it sets the line offset to the cursor's line",
				c: &Cursor{
					col:        1,
					line:       10,
					lineOffset: 15,
				},
				wantCursor: &Cursor{
					col:        1,
					line:       10,
					lineOffset: 9,
				},
			},
			{
				name: "when the cursor's zero-indexed line is equal to the current line offset, " +
					"it does nothing",
				c: &Cursor{
					col:        1,
					line:       16,
					lineOffset: 15,
				},
				wantCursor: &Cursor{
					col:        1,
					line:       16,
					lineOffset: 15,
				},
			},
			{
				name: "when the cursor's zero-indexed line is below the current line offset, " +
					"it does nothing",
				c: &Cursor{
					col:        1,
					line:       20,
					lineOffset: 15,
				},
				wantCursor: &Cursor{
					col:        1,
					line:       20,
					lineOffset: 15,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				tc.c.scroll(20, 20)
				if !reflect.DeepEqual(tc.c, tc.wantCursor) {
					t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
				}
			})
		}
	})

	t.Run("scroll down", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name       string
			c          *Cursor
			wantCursor *Cursor
		}{
			{
				name: "when the cursor's zero-indexed line is below the current screen, " +
					"it updates the offset so the cursor is on the last line of the screen",
				c: &Cursor{
					col:        1,
					line:       20,
					lineOffset: 5,
				},
				wantCursor: &Cursor{
					col:        1,
					line:       20,
					lineOffset: 15,
				},
			},
			{
				name: "when the cursor's zero-indexed line is the last line of the screen" +
					"it does nothing",
				c: &Cursor{
					col:        1,
					line:       20,
					lineOffset: 15,
				},
				wantCursor: &Cursor{
					col:        1,
					line:       20,
					lineOffset: 15,
				},
			},
			{
				name: "when the cursor's zero-indexed line is above the last line of the screen, " +
					"it does nothing",
				c: &Cursor{
					col:        1,
					line:       8,
					lineOffset: 5,
				},
				wantCursor: &Cursor{
					col:        1,
					line:       8,
					lineOffset: 5,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				tc.c.scroll(5, 5)
				if !reflect.DeepEqual(tc.c, tc.wantCursor) {
					t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
				}
			})
		}
	})

	t.Run("scroll left", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name       string
			c          *Cursor
			wantCursor *Cursor
		}{
			{
				name: "when the cursor is left of the left margin and is less than " +
					"defaultCursorMargin characters from the start of the line, it updates " +
					"the column offset to the start of the line",
				c: &Cursor{
					col:       defaultCursorMargin - 2,
					line:      1,
					colOffset: 5,
				},
				wantCursor: &Cursor{
					col:       defaultCursorMargin - 2,
					line:      1,
					colOffset: 0,
				},
			},
			{
				name: "when the cursor is left of the left margin and is defaultCursorMargin" +
					"characters from the start of the line, it updates the column offset to the " +
					"start of the line",
				c: &Cursor{
					col:       defaultCursorMargin + 1, // col is zero-indexed
					line:      1,
					colOffset: 5,
				},
				wantCursor: &Cursor{
					col:       defaultCursorMargin + 1,
					line:      1,
					colOffset: 0,
				},
			},
			{
				name: "when the cursor is left of the left margin and is more than " +
					"defaultCursorMargin characters from the start of the line, it updates " +
					"the column offset to show defaultCursorMargin characters before the cursor",
				c: &Cursor{
					col:       defaultCursorMargin + 2,
					line:      1,
					colOffset: 5,
				},
				wantCursor: &Cursor{
					col:       defaultCursorMargin + 2,
					line:      1,
					colOffset: 1,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				tc.c.scroll(10, 10)
				if !reflect.DeepEqual(tc.c, tc.wantCursor) {
					t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
				}
			})
		}
	})

	t.Run("scroll right", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name       string
			c          *Cursor
			wantCursor *Cursor
		}{
			{
				name: "when the cursor is left of the right margin, it does nothing",
				c: &Cursor{
					col:       5,
					line:      1,
					colOffset: 0,
				},
				wantCursor: &Cursor{
					col:       5,
					line:      1,
					colOffset: 0,
				},
			},
			{
				name: "when the cursor is on the right margin, it updates the column offset to " +
					"place the cursor in the rightmost column",
				c: &Cursor{
					col:       20,
					line:      1,
					colOffset: 0,
				},
				wantCursor: &Cursor{
					col:       20,
					line:      1,
					colOffset: 10,
				},
			},
			{
				name: "when the cursor is right of the right margin, it updates the column offset " +
					"to place the cursor in the rightmost column",
				c: &Cursor{
					col:       25,
					line:      1,
					colOffset: 0,
				},
				wantCursor: &Cursor{
					col:       25,
					line:      1,
					colOffset: 15,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				tc.c.scroll(10, 10)
				if !reflect.DeepEqual(tc.c, tc.wantCursor) {
					t.Errorf("expected cursor to be %+v, got %+v", tc.wantCursor, tc.c)
				}
			})
		}
	})
}
