package encryptor

type PositionBased struct {
	KeyGenerator XORKeyGenerator
	position     uint32
}

func NewPositionBasedEncryptor(gen XORKeyGenerator, pos uint32) *PositionBased {
	return &PositionBased{
		KeyGenerator: gen,
		position:     pos,
	}
}

// implements cipher.Stream
func (e *PositionBased) XORKeyStream(dst, src []byte) {
	blockSize := e.KeyGenerator.BlockSize()

	var key []byte
	keyOffset := blockSize
	for i := uint32(0); i < uint32(len(src)); i++ {
		if keyOffset == blockSize {
			key = e.KeyGenerator.Key((e.position + i) / blockSize)
			keyOffset = 0
		}
		dst[i] = src[i] ^ key[keyOffset]
		keyOffset = keyOffset + 1
	}
	e.position += uint32(len(src))
}

func (e *PositionBased) Reset() {
	e.position = 0
}
