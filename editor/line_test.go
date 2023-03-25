package editor

import (
	"reflect"
	"testing"
)

func Test_Line_RuneLen(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		l    *Line
		want int
	}{
		{
			name: "nil",
			l:    nil,
			want: 0,
		},
		{
			name: "empty",
			l:    newLine(),
			want: 0,
		},
		{
			name: "non-empty",
			l:    newLineFromString("hello"),
			want: 5,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.l.RuneLen(); got != tc.want {
				t.Errorf("Line.RuneLen() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_Line_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		l    *Line
		want string
	}{
		{
			name: "nil",
			l:    nil,
			want: "",
		},
		{
			name: "empty",
			l:    newLine(),
			want: "",
		},
		{
			name: "non-empty",
			l:    newLineFromString("hello"),
			want: "hello",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.l.String(); got != tc.want {
				t.Errorf("Line.String() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_Line_Runes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		l    *Line
		want []rune
	}{
		{
			name: "nil",
			l:    nil,
			want: nil,
		},
		{
			name: "empty",
			l:    newLine(),
			want: make([]rune, 0, lineRunesToPreallocate),
		},
		{
			name: "non-empty",
			l:    newLineFromString("hello"),
			want: []rune("hello"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.l.Runes(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Line.Runes() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_newLine(t *testing.T) {
	t.Parallel()

	l := newLine()
	if len(l.runes) != 0 {
		t.Errorf("expected len(runes) 0, got %d", len(l.runes))
	}
	if cap(l.runes) != lineRunesToPreallocate {
		t.Errorf("expected cap(runes) %d, got %d", lineRunesToPreallocate, cap(l.runes))
	}
}

func Test_newLineFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		s    string
		want *Line
	}{
		{
			name: "when the string is empty " +
				"it returns an empty line",
			s:    "",
			want: newLine(),
		},
		{
			name: "when the string contains no tabs " +
				"it is converted to runes unchanged",
			s: "hello",
			want: &Line{
				runes: []rune("hello"),
			},
		},
		{
			name: "when a tab occurs at the start of a tab stop " +
				"it is replaced by tabStop spaces",
			s: "hell\tworld",
			want: &Line{
				runes: []rune("hell    world"),
			},
		},
		{
			name: "when a tab occurs n characters into a tab stop " +
				"it is replaced by tabStop-n spaces",
			s: "hello\tworld",
			want: &Line{
				runes: []rune("hello   world"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := newLineFromString(tc.s); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %+v, got %+v", tc.want, got)
			}
		})
	}
}

func Test_newLineFromRunes(t *testing.T) {
	// Test that asserts that the runes passed as an argument to the function are used as the runes field in the returned line.
	t.Parallel()

	testCases := []struct {
		name string
		r    []rune
		want *Line
	}{
		{
			name: "when the runes are empty " +
				"it returns an empty line",
			r:    nil,
			want: &Line{runes: nil},
		},
		{
			name: "when the runes are not empty " +
				"it returns a line with the runes",
			r:    []rune("hello"),
			want: &Line{runes: []rune("hello")},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := newLineFromRunes(tc.r); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %+v, got %+v", tc.want, got)
			}
		})
	}

}

func Test_Line_insertRuneAt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		l    *Line
		r    rune
		i    int
		want *Line
	}{
		{
			name: "when the line is empty " +
				"it inserts the rune at the start",
			l:    newLine(),
			r:    'a',
			i:    0,
			want: newLineFromString("a"),
		},
		{
			name: "when the line is not empty " +
				"it inserts the rune at the specified index",
			l:    newLineFromString("hello"),
			r:    'a',
			i:    2,
			want: newLineFromString("heallo"),
		},
		{
			name: "when the index is < 0 " +
				"it inserts the rune at the end",
			l:    newLineFromString("hello"),
			r:    'a',
			i:    -1,
			want: newLineFromString("helloa"),
		},
		{
			name: "when the index is > len " +
				"it inserts the rune at the end",
			l:    newLineFromString("hello"),
			r:    'a',
			i:    10,
			want: newLineFromString("helloa"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.l.insertRuneAt(tc.r, tc.i)

			if !reflect.DeepEqual(tc.l, tc.want) {
				t.Errorf("expected %#v, got %#v", tc.want, tc.l)
			}
		})
	}
}

func Test_Line_appendRune(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		l    *Line
		r    rune
		want *Line
	}{
		{
			name: "when the line is empty " +
				"it appends the rune",
			l:    newLine(),
			r:    'a',
			want: newLineFromString("a"),
		},
		{
			name: "when the line is not empty " +
				"it appends the rune",
			l:    newLineFromString("hello"),
			r:    'a',
			want: newLineFromString("helloa"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.l.appendRune(tc.r)

			if !reflect.DeepEqual(tc.l, tc.want) {
				t.Errorf("expected %#v, got %#v", tc.want, tc.l)
			}
		})
	}
}

func Test_Line_clear(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		l    *Line
		want *Line
	}{
		{
			name: "when the line is empty " +
				"it clears the line",
			l:    newLine(),
			want: newLine(),
		},
		{
			name: "when the line is not empty " +
				"it clears the line",
			l:    newLineFromString("hello"),
			want: newLine(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.l.clear()

			if !reflect.DeepEqual(tc.l, tc.want) {
				t.Errorf("expected %#v, got %#v", tc.want, tc.l)
			}
		})
	}
}

func Test_Line_deleteRuneAt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		l    *Line
		i    int
		want *Line
	}{
		{
			name: "when the line is empty " +
				"it does nothing",
			l:    newLine(),
			i:    0,
			want: newLine(),
		},
		{
			name: "when the index is < 0 " +
				"it deletes the last rune",
			l:    newLineFromString("hello"),
			i:    -1,
			want: newLineFromString("hell"),
		},
		{
			name: "when the index is > len " +
				"it deletes the last rune",
			l:    newLineFromString("hello"),
			i:    10,
			want: newLineFromString("hell"),
		},
		{
			name: "when the index is valid " +
				"it deletes the rune at the index",
			l:    newLineFromString("hello"),
			i:    2,
			want: newLineFromString("helo"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.l.deleteRuneAt(tc.i)

			if !reflect.DeepEqual(tc.l, tc.want) {
				t.Errorf("expected %#v, got %#v", tc.want, tc.l)
			}
		})
	}
}

func Test_Line_deleteLastRune(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		l    *Line
		want *Line
	}{
		{
			name: "when the line is empty " +
				"it does nothing",
			l:    newLine(),
			want: newLine(),
		},
		{
			name: "when the line is not empty " +
				"it deletes the last rune",
			l:    newLineFromString("hello"),
			want: newLineFromString("hell"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.l.deleteLastRune()

			if !reflect.DeepEqual(tc.l, tc.want) {
				t.Errorf("expected %#v, got %#v", tc.want, tc.l)
			}
		})
	}
}

func Test_Line_append(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		l     *Line
		other *Line
		want  *Line
	}{
		{
			name: "when the line is empty " +
				"it appends the other line",
			l:     newLine(),
			other: newLineFromString("hello"),
			want:  newLineFromString("hello"),
		},
		{
			name: "when the other line is empty " +
				"it does nothing",
			l:     newLineFromString("hello"),
			other: newLine(),
			want:  newLineFromString("hello"),
		},
		{
			name: "when the line is not empty " +
				"it appends the other line",
			l:     newLineFromString("hello"),
			other: newLineFromString("world"),
			want:  newLineFromString("helloworld"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.l.append(tc.other)

			if !reflect.DeepEqual(tc.l, tc.want) {
				t.Errorf("expected %#v, got %#v", tc.want, tc.l)
			}
		})
	}
}
