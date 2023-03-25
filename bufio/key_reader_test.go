package bufio

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

// MockReader is a mock io.Reader.
type MockReader struct {
	readFunc func(p []byte) (n int, err error)
}

// Read satisfies the io.Reader interface.
func (r *MockReader) Read(p []byte) (n int, err error) {
	return r.readFunc(p)
}

func Test_NewKeyReader(t *testing.T) {
	t.Parallel()

	r := strings.NewReader("hello")
	kr := NewKeyReader(r, 5)
	if kr.r != r {
		t.Errorf("NewKeyReader() = %+v, want %+v", kr.r, r)
	}
	if len(kr.keyBuf) != 5 {
		t.Errorf("NewKeyReader() = %+v, want %+v", len(kr.keyBuf), 5)
	}
}

func Test_KeyReader_ReadKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		r       *MockReader
		want    []byte
		wantErr error
	}{
		{
			name: "when the call to r.Read succeeds " +
				"ReadKey returns the bytes read from the underlying reader",
			r: &MockReader{
				readFunc: func(p []byte) (n int, err error) {
					n = copy(p, "hello")
					return
				},
			},
			want: []byte("hello"),
		},
		{
			name: "when the call to r.Read fails " +
				"ReadKey returns the error returned by the underlying reader",
			r: &MockReader{
				readFunc: func(p []byte) (n int, err error) {
					return 0, io.EOF
				},
			},
			wantErr: io.EOF,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kr := &KeyReader{
				r:      tc.r,
				keyBuf: make([]byte, 5),
			}
			got, err := kr.ReadKey()
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("KeyReader.ReadKey() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("KeyReader.ReadKey() = %v, want %v", got, tc.want)
			}
		})
	}
}
