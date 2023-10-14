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

	body, err := io.ReadAll(r)
	if err != nil {
		return errors.New("read body: " + err.Error())
	}

	sum := md5.Sum(body)

	if !bytes.Equal(sum[:], header.Checksum[:]) {
		return errors.New("checksum mismatch")
	}

	switch header.Encryption {
	case EncryptNone:
	case EncryptStream:
		enc := encryptor.NewBlock512KeyGenerator(key)
		stream := encryptor.NewPositionBasedEncryptor(enc, 0)
		stream.XORKeyStream(body, body)
	default:
		return fmt.Errorf("unknown encryption mode: %d", header.Encryption)
	}

	_, err = out.Write(body)

	return err
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
