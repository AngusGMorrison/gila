package editor

import "io"

// KeyReader reads a single keystroke or chord from input and returns its raw bytes.
type KeyReader interface {
	ReadKey() ([]byte, error)
}

// ZeroAllocationKeyReader avoids allocations when reading keys from input by maintaining a buffer,
// keyBuf, that is returned to the caller by ReadKey and shared between ReadKey calls.
type ZeroAllocationKeyReader struct {
	r      io.Reader
	keyBuf []byte
}

func NewZeroAllocationKeyReader(r io.Reader) *ZeroAllocationKeyReader {
	return &ZeroAllocationKeyReader{
		r:      r,
		keyBuf: make([]byte, readMaxBytes),
	}
}

// ReadKey attempts to read the bytes corresponding to a keypress or chord into keyBuf. When the
// underlying reader is a terminal in raw mode, ReadKey will block until at least one byte is read.
// The return value is a slice containing the n bytes read from r. This slice shares the same
// underlying memory as keyBuf, making it unsafe to reuse between calls to Read.
func (kr *ZeroAllocationKeyReader) ReadKey() ([]byte, error) {
	n, err := kr.r.Read(kr.keyBuf)
	if err != nil {
		return nil, err
	}
	return kr.keyBuf[:n], nil
}
