package bufio

import (
	"bufio"
	"reflect"
	"testing"

	"github.com/angusgmorrison/gila/escseq"
)

// MockWriter is a mock io.Writer.
type MockWriter struct {
	writeFunc func(p []byte) (n int, err error)
}

// Write satisfies the io.Writer interface.
func (w *MockWriter) Write(p []byte) (n int, err error) {
	return w.writeFunc(p)
}

func Test_NewTerminalWriter(t *testing.T) {
	t.Parallel()

	w := &MockWriter{}
	tw := NewTerminalWriter(w)
	want := &TerminalWriter{
		w: bufio.NewWriterSize(w, defaultBufferBytes),
	}
	if !reflect.DeepEqual(tw, want) {
		t.Errorf("expected %+v, want %+v", tw.w, w)
	}
}

// Table-driven tests for TerminalWriter_WriteEscapeSequence, testing the written output using MockWriter.
// 1. When no formatting directives are present, the escape sequence is written.
// 2. When formatting directives are present, the escape sequence is formatted before being written.
func Test_TerminalWriter_WriteEscapeSequence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		esc  escseq.EscSeq
		args []interface{}
		want string
	}{
		{
			name: "no formatting directives",
			esc:  escseq.EscSeq("test"),
			args: nil,
			want: "test",
		},
		{
			name: "formatting directives",
			esc:  escseq.EscSeq("test %s"),
			args: []interface{}{"test"},
			want: "test test",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got string
			w := &MockWriter{
				writeFunc: func(p []byte) (n int, err error) {
					got = string(p)
					return len(p), nil
				},
			}
			tw := NewTerminalWriter(w)
			_, err := tw.WriteEscapeSequence(tc.esc, tc.args...)
			if err != nil {
				t.Errorf("unexpected error writing escape sequence: %#v", err)
			}
			err = tw.Flush()
			if err != nil {
				t.Errorf("unexpected error flushing buffer: %#v", err)
			}
			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
