package editor

import (
	"strings"
	"unicode/utf8"
)

const (
	tabStop                = 4
	lineRunesToPreallocate = 128
)

// Line represents a single line of text.
// TODO: []rune represents the cleanest way of handling UTF-8-encoded
// strings, but requires allocations to copy from string when reading
// input, and to string when writing output. Can this be avoided?
type Line struct {
	raw   string
	runes []rune
}

// RuneLen returns the length of the line as it appears to the user.
func (l *Line) RuneLen() int {
	if l == nil {
		return 0
	}
	return len(l.runes)
}

// String returns the rendered view of the line.
func (l *Line) String() string {
	return string(l.runes)
}

func (l *Line) Runes() []rune {
	return l.runes
}

func newLine() *Line {
	return &Line{
		runes: make([]rune, 0, lineRunesToPreallocate),
	}
}

func newLineFromRunes(r []rune) *Line {
	return &Line{
		runes: r,
	}
}

func newLineFromString(s string) *Line {
	// Replace tabs with spaces to override terminal tab stop setting.
	tabs := strings.Count(s, "\t")
	spaces := tabs * (tabStop - 1) // the additional spaces required to replace tabs
	runes := utf8.RuneCountInString(s)
	render := make([]rune, 0, runes+spaces)
	for _, r := range s {
		if r == '\t' {
			render = append(render, ' ')
			for len(render)%tabStop != 0 {
				render = append(render, ' ')
			}
		} else {
			render = append(render, r)
		}
	}

	return &Line{
		raw:   s,
		runes: render,
	}
}

func (l *Line) insertRuneAt(r rune, i int) {
	if i < 0 || i > l.RuneLen() {
		i = l.RuneLen()
	}
	l.runes = append(l.runes[:i], append([]rune{r}, l.runes[i:]...)...)
}

func (l *Line) appendRune(r rune) {
	l.runes = append(l.runes, r)
}

func (l *Line) clear() {
	l.raw = ""
	l.runes = l.runes[:0]
}

func (l *Line) deleteRuneAt(i int) {
	if i < 0 || i >= l.RuneLen() {
		i = l.RuneLen() - 1
	}
	l.runes = append(l.runes[:i], l.runes[i+1:]...)
}

func (l *Line) append(other *Line) {
	l.runes = append(l.runes, other.runes...)
}
