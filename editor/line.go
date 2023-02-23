package editor

import (
	"strings"
	"unicode/utf8"
)

const tabStop = 4

// Line represents a single line of text.
// TODO: []rune represents the cleanest way of handling UTF-8-encoded
// strings, but requires allocations to copy from string when reading
// input, and to string when writing output. Can this be avoided?
type Line struct {
	raw    string
	render []rune
}

// RenderLen returns the length of the line as it appears to the user.
func (l *Line) RenderLen() uint {
	if l == nil {
		return 0
	}
	return uint(len(l.render))
}

// String returns the rendered view of the line.
func (l *Line) String() string {
	return string(l.render)
}

func newLine(s string) *Line {
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
		raw:    s,
		render: render,
	}
}

// func (l *line) update() {
// 	if cap(l.render) < len(l.raw) {
// 		l.render = make([]rune, len(l.raw))
// 		copy(l.render, l.raw)
// 		return
// 	}

// 	if len(l.render) < len(l.raw) {
// 		copy(l.render, l.raw)
// 		l.render = append(l.render, l.raw[len(l.render):]...)
// 		return
// 	}

// 	copy(l.render, l.raw)
// 	l.render = l.render[:len(l.raw)]
// }
