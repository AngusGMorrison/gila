package bufio

import (
	"io"

	"github.com/angusgmorrison/gila/editor"
)

// KeyReader satisfies editor.KeyReader. It avoids allocations when reading keys
// from input by maintaining a buffer, keyBuf, that is returned to the caller by
// ReadKey and shared between ReadKey calls.
type KeyReader struct {
	r      io.Reader
	keyBuf []byte
}

var _ editor.KeyReader = (*KeyReader)(nil)

// NewKeyReader returns a *KeyReader with an key buffer of len maxKeyBytes. This
// is the maximum size of keypress it must be able to read in bytes.
func NewKeyReader(r io.Reader, maxKeyBytes int) *KeyReader {
	return &KeyReader{
		r:      r,
		keyBuf: make([]byte, maxKeyBytes),
	}
}

// ReadKey attempts to read the bytes corresponding to a keypress or chord into
// keyBuf. When the underlying reader is a terminal in raw mode, ReadKey will
// block until at least one byte is read. The return value is a slice containing
// the n bytes read from r. This slice shares the same underlying memory as
// keyBuf, making it unsafe to reuse between calls to Read.
func (kr *KeyReader) ReadKey() ([]byte, error) {
	n, err := kr.r.Read(kr.keyBuf)
	if err != nil {
		return nil, err
	}
	return kr.keyBuf[:n], nil
}
