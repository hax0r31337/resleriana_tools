package encryptor

import (
	"crypto/sha512"
	"encoding/binary"
	"math/bits"
	"unsafe"
)

type XORKeyGenerator interface {
	BlockSize() uint32
	Key(block uint32) []byte
}

// a modified version of chacha
type Block512KeyGenerator struct {
	key   [128]uint32 // pre-allocate key array to improve performance
	state [16]uint32
	hash2 []byte
}

// key: "{bundlename}-{size}-{hash}-{crc}"
func NewBlock512KeyGenerator(key []byte) *Block512KeyGenerator {
	h1 := sha512.Sum512(key)
	h2 := sha512.Sum512(h1[:])

	g := &Block512KeyGenerator{
		hash2: h2[:],
	}

	// "expand 32-byte k"
	g.state[0] = 0x61707865
	g.state[1] = 0x3320646e
	g.state[2] = 0x79622d32
	g.state[3] = 0x6b206574

	innerKey := unsafe.Slice((*byte)(unsafe.Pointer(&g.state[4])), 32)
	copy(innerKey, h1[:32])

	return g
}

func (Block512KeyGenerator) BlockSize() uint32 {
	return 512
}

func (g *Block512KeyGenerator) Key(block uint32) []byte {
	// generate 12-bit nonce
	hashPart1 := binary.LittleEndian.Uint32(g.hash2[(block%0xd)|0x30:])
	hashPart2 := binary.LittleEndian.Uint32(g.hash2[block/0xd%0xd:])
	hashPart3 := binary.LittleEndian.Uint32(g.hash2[(block/0xa9%0xd)|0x10:])
	hashPart4 := binary.LittleEndian.Uint32(g.hash2[(block/0x895%0xd)|0x20:])

	hashPartRotated := bits.RotateLeft32(hashPart1, 2*int(block/0xa9)-28*int((0x24924925*int64(block/0x152))>>32)) ^ bits.RotateLeft32(hashPart2, int(3*(block/0x93e)%0x1b))

	g.state[13] = hashPartRotated
	g.state[14] = hashPartRotated ^ hashPart3
	g.state[15] = hashPartRotated ^ hashPart3 ^ hashPart4

	// do 8 times HChaCha for 512-byte key

	g.state[12] = block + 1
	g.generateBlock(12, 0x00)

	g.state[12]++
	g.generateBlock(8, 0x10)

	g.state[12]++
	g.generateBlock(8, 0x20)

	g.state[12]++
	g.generateBlock(8, 0x30)

	g.state[12]++
	g.generateBlock(4, 0x40)

	g.state[12]++
	g.generateBlock(4, 0x50)

	g.state[12]++
	g.generateBlock(4, 0x60)

	g.state[12]++
	g.generateBlock(4, 0x70)

	return unsafe.Slice((*byte)(unsafe.Pointer(&g.key[0])), len(g.key)*4)
}

func (g *Block512KeyGenerator) generateBlock(rounds int, offset int) {
	if rounds%2 != 0 {
		panic("rounds must be multiple of 2")
	} else if offset%16 != 0 {
		panic("offset must be multiple of 16")
	}

	if offset == 0 {
		for i := 0; i < len(g.state); i++ {
			g.key[offset+i] = g.state[i]
		}
	} else {
		// xor with previous generated block
		for i := 0; i < len(g.state); i++ {
			g.key[offset+i] = g.key[offset+i-16] ^ g.state[i]
		}
	}

	mix := g.key[offset : offset+16]

	// no modification between original HChaCha specification
	for i := 0; i < rounds; i += 2 {
		quarterRound(mix, 0, 4, 8, 12)
		quarterRound(mix, 1, 5, 9, 13)
		quarterRound(mix, 2, 6, 10, 14)
		quarterRound(mix, 3, 7, 11, 15)

		quarterRound(mix, 0, 5, 10, 15)
		quarterRound(mix, 1, 6, 11, 12)
		quarterRound(mix, 2, 7, 8, 13)
		quarterRound(mix, 3, 4, 9, 14)
	}

	// adds original state back to generated xor key as a part of HChaCha specification
	if offset == 0 {
		for i := 0; i < len(g.state); i++ {
			g.key[offset+i] += g.state[i]
		}
	} else {
		for i := 0; i < len(g.state); i++ {
			g.key[offset+i] += g.key[offset+i-16] ^ g.state[i]
		}
	}
}

func quarterRound(output []uint32, a, b, c, d int) {
	output[a] += output[b]
	output[d] = bits.RotateLeft32(output[d]^output[a], 16)

	output[c] += output[d]
	output[b] = bits.RotateLeft32(output[b]^output[c], 12)

	output[a] += output[b]
	output[d] = bits.RotateLeft32(output[d]^output[a], 8)

	output[c] += output[d]
	output[b] = bits.RotateLeft32(output[b]^output[c], 7)
}
