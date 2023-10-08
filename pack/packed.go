package pack

import (
	"aktsk/encryptor"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const (
	EncryptNone = iota
	// encrypted with Block512KeyGenerator, which is a modified version of chacha
	EncryptStream
)

type PackedABHeader_v1 struct {
	Reserved   uint16
	Encryption uint32
	Checksum   [16]byte
}

// ReadFrom implements io.ReaderFrom
func (h *PackedABHeader_v1) ReadFrom(r io.Reader) (int64, error) {
	err := binary.Read(r, binary.LittleEndian, &h.Reserved)
	if err != nil {
		return 0, errors.New("read version: " + err.Error())
	}
	err = binary.Read(r, binary.LittleEndian, &h.Encryption)
	if err != nil {
		return 0, errors.New("read version: " + err.Error())
	}
	return 0, mustRead(r, h.Checksum[:])
}

func ReadPackedAB(r io.Reader, out io.Writer, key []byte) error {
	magic := make([]byte, 4)
	err := mustRead(r, magic)
	if err != nil {
		return errors.New("read magic: " + err.Error())
	}
	if string(magic) != "Aktk" {
		return fmt.Errorf("unexpected magic, expect %s but got %s", hex.EncodeToString([]byte("Aktk")), hex.EncodeToString(magic))
	}

	var version uint16
	err = binary.Read(r, binary.LittleEndian, &version)
	if err != nil {
		return errors.New("read version: " + err.Error())
	}

	if version != 0x01 {
		return fmt.Errorf("unsupported version: %d", version)
	}

	header := &PackedABHeader_v1{}
	_, err = header.ReadFrom(r)
	if err != nil {
		return errors.New("read header: " + err.Error())
	}

	hasher := md5.New()
	r = encryptor.NewHashReader(r, hasher)

	switch header.Encryption {
	case EncryptNone:
	case EncryptStream:
		enc := encryptor.NewBlock512KeyGenerator(key)
		stream := encryptor.NewPositionBasedEncryptor(enc, 0)
		r = encryptor.NewStreamCipherReader(r, stream)
	default:
		return fmt.Errorf("unknown encryption mode: %d", header.Encryption)
	}

	_, err = io.Copy(out, r)
	if err != nil {
		return err
	}

	if !bytes.Equal(hasher.Sum(nil), header.Checksum[:]) {
		println(hex.EncodeToString(hasher.Sum(nil)))
		return errors.New("checksum mismatch")
	}

	return nil
}

func mustRead(r io.Reader, arr []byte) error {
	n, err := r.Read(arr)
	if err != nil {
		return err
	} else if n < len(arr) {
		return fmt.Errorf("insufficient data, expect %d but got %d", len(arr), n)
	}
	return nil
}
