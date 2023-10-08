package encryptor

import (
	"crypto/cipher"
	"hash"
	"io"
)

type StreamCipherReader struct {
	io.Reader

	stream cipher.Stream
}

func NewStreamCipherReader(r io.Reader, stream cipher.Stream) *StreamCipherReader {
	return &StreamCipherReader{
		Reader: r,
		stream: stream,
	}
}

// Read implements io.Reader
func (r *StreamCipherReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err != nil {
		return n, err
	}

	r.stream.XORKeyStream(p, p)

	return n, err
}

type HasherReader struct {
	io.Reader

	hash hash.Hash
}

func NewHashReader(r io.Reader, hasher hash.Hash) *HasherReader {
	return &HasherReader{
		Reader: r,
		hash:   hasher,
	}
}

// Read implements io.Reader
func (r *HasherReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err != nil {
		return n, err
	}

	r.hash.Write(p[:n])

	return n, err
}
