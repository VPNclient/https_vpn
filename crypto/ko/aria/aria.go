package aria

import (
	"crypto/cipher"
	"fmt"
)

// ariaCipher implements the cipher.Block interface for ARIA
type ariaCipher struct {
	rounds int
	ek     []byte // encryption round keys
	dk     []byte // decryption round keys
}

// NewCipher creates an ARIA cipher with automatic key size detection
func NewCipher(key []byte) (cipher.Block, error) {
	switch len(key) {
	case KeySize128:
		return NewCipher128(key)
	case KeySize192:
		return NewCipher192(key)
	case KeySize256:
		return NewCipher256(key)
	default:
		return nil, fmt.Errorf("aria: invalid key size %d (must be 16, 24, or 32)", len(key))
	}
}

// NewCipher128 creates an ARIA-128 cipher
func NewCipher128(key []byte) (cipher.Block, error) {
	if len(key) != KeySize128 {
		return nil, fmt.Errorf("aria: invalid key size %d (must be 16)", len(key))
	}
	return newAriaCipher(key, Rounds128)
}

// NewCipher192 creates an ARIA-192 cipher
func NewCipher192(key []byte) (cipher.Block, error) {
	if len(key) != KeySize192 {
		return nil, fmt.Errorf("aria: invalid key size %d (must be 24)", len(key))
	}
	return newAriaCipher(key, Rounds192)
}

// NewCipher256 creates an ARIA-256 cipher
func NewCipher256(key []byte) (cipher.Block, error) {
	if len(key) != KeySize256 {
		return nil, fmt.Errorf("aria: invalid key size %d (must be 32)", len(key))
	}
	return newAriaCipher(key, Rounds256)
}

func newAriaCipher(key []byte, rounds int) (*ariaCipher, error) {
	c := &ariaCipher{
		rounds: rounds,
		ek:     make([]byte, (rounds+1)*BlockSize),
		dk:     make([]byte, (rounds+1)*BlockSize),
	}
	c.expandKey(key)
	return c, nil
}

// BlockSize returns the ARIA block size
func (c *ariaCipher) BlockSize() int {
	return BlockSize
}

// Encrypt encrypts a single block
func (c *ariaCipher) Encrypt(dst, src []byte) {
	if len(src) < BlockSize {
		panic("aria: input not full block")
	}
	if len(dst) < BlockSize {
		panic("aria: output not full block")
	}
	c.encrypt(dst, src)
}

// Decrypt decrypts a single block
func (c *ariaCipher) Decrypt(dst, src []byte) {
	if len(src) < BlockSize {
		panic("aria: input not full block")
	}
	if len(dst) < BlockSize {
		panic("aria: output not full block")
	}
	c.decrypt(dst, src)
}

// expandKey performs ARIA key expansion
func (c *ariaCipher) expandKey(key []byte) {
	var w [4][16]byte
	var kl, kr [16]byte

	keyLen := len(key)

	// Initialize KL and KR from the key
	copy(kl[:], key)
	if keyLen > 16 {
		copy(kr[:], key[16:])
	}

	// Determine CK index based on key size
	ckIdx := 0
	if keyLen == KeySize192 {
		ckIdx = 1
	} else if keyLen == KeySize256 {
		ckIdx = 2
	}

	// Generate W0, W1, W2, W3
	// W0 = KL
	copy(w[0][:], kl[:])

	// W1 = FO(W0, CK1) XOR KR
	c.fo(w[1][:], w[0][:], ck[ckIdx][:])
	xorBlock(w[1][:], w[1][:], kr[:])

	// W2 = FE(W1, CK2) XOR W0
	c.fe(w[2][:], w[1][:], ck[(ckIdx+1)%3][:])
	xorBlock(w[2][:], w[2][:], w[0][:])

	// W3 = FO(W2, CK3) XOR W1
	c.fo(w[3][:], w[2][:], ck[(ckIdx+2)%3][:])
	xorBlock(w[3][:], w[3][:], w[1][:])

	// Generate encryption round keys
	c.generateRoundKeys(w[:])
}

// generateRoundKeys generates round keys from W
func (c *ariaCipher) generateRoundKeys(w [][16]byte) {
	// Generate encryption keys
	// ek1 = W0 XOR (W1 >>> 19)
	c.genRoundKey(c.ek[0:16], w[0][:], w[1][:], 19)
	// ek2 = W1 XOR (W2 >>> 19)
	c.genRoundKey(c.ek[16:32], w[1][:], w[2][:], 19)
	// ek3 = W2 XOR (W3 >>> 19)
	c.genRoundKey(c.ek[32:48], w[2][:], w[3][:], 19)
	// ek4 = W3 XOR (W0 >>> 19)
	c.genRoundKey(c.ek[48:64], w[3][:], w[0][:], 19)
	// ek5 = W0 XOR (W1 >>> 31)
	c.genRoundKey(c.ek[64:80], w[0][:], w[1][:], 31)
	// ek6 = W1 XOR (W2 >>> 31)
	c.genRoundKey(c.ek[80:96], w[1][:], w[2][:], 31)
	// ek7 = W2 XOR (W3 >>> 31)
	c.genRoundKey(c.ek[96:112], w[2][:], w[3][:], 31)
	// ek8 = W3 XOR (W0 >>> 31)
	c.genRoundKey(c.ek[112:128], w[3][:], w[0][:], 31)
	// ek9 = W0 XOR (W1 <<< 61)
	c.genRoundKey(c.ek[128:144], w[0][:], w[1][:], 67) // 128-61=67 right rotation
	// ek10 = W1 XOR (W2 <<< 61)
	c.genRoundKey(c.ek[144:160], w[1][:], w[2][:], 67)
	// ek11 = W2 XOR (W3 <<< 61)
	c.genRoundKey(c.ek[160:176], w[2][:], w[3][:], 67)
	// ek12 = W3 XOR (W0 <<< 61)
	c.genRoundKey(c.ek[176:192], w[3][:], w[0][:], 67)

	if c.rounds >= Rounds192 {
		// ek13 = W0 XOR (W1 <<< 31)
		c.genRoundKey(c.ek[192:208], w[0][:], w[1][:], 97) // 128-31=97
	}
	if c.rounds >= Rounds256 {
		// ek14 = W1 XOR (W2 <<< 31)
		c.genRoundKey(c.ek[208:224], w[1][:], w[2][:], 97)
		// ek15 = W2 XOR (W3 <<< 31)
		c.genRoundKey(c.ek[224:240], w[2][:], w[3][:], 97)
	}
	// Last round key
	switch c.rounds {
	case Rounds128:
		c.genRoundKey(c.ek[192:208], w[0][:], w[1][:], 97)
	case Rounds192:
		c.genRoundKey(c.ek[224:240], w[1][:], w[2][:], 97)
	case Rounds256:
		c.genRoundKey(c.ek[256:272], w[3][:], w[0][:], 97)
	}

	// Generate decryption keys (reverse order with diffusion layer)
	numKeys := c.rounds + 1
	// First decryption key = last encryption key
	copy(c.dk[0:16], c.ek[(numKeys-1)*16:numKeys*16])
	// Middle keys with diffusion layer
	for i := 1; i < numKeys-1; i++ {
		c.diffusionA(c.dk[i*16:(i+1)*16], c.ek[(numKeys-1-i)*16:(numKeys-i)*16])
	}
	// Last decryption key = first encryption key
	copy(c.dk[(numKeys-1)*16:numKeys*16], c.ek[0:16])
}

// genRoundKey generates a single round key
func (c *ariaCipher) genRoundKey(dst, w1, w2 []byte, rot int) {
	var rotated [16]byte
	rotateRight128(rotated[:], w2, rot)
	xorBlock(dst, w1, rotated[:])
}

// rotateRight128 rotates a 128-bit value right by n bits
func rotateRight128(dst, src []byte, n int) {
	n = n % 128
	if n == 0 {
		copy(dst, src)
		return
	}

	byteShift := n / 8
	bitShift := n % 8

	for i := 0; i < 16; i++ {
		srcIdx1 := (i + byteShift) % 16
		srcIdx2 := (i + byteShift + 1) % 16
		if bitShift == 0 {
			dst[i] = src[srcIdx1]
		} else {
			dst[i] = (src[srcIdx1] >> bitShift) | (src[srcIdx2] << (8 - bitShift))
		}
	}
}

// xorBlock XORs two 16-byte blocks
func xorBlock(dst, a, b []byte) {
	for i := 0; i < 16; i++ {
		dst[i] = a[i] ^ b[i]
	}
}

// fo is the odd round function
func (c *ariaCipher) fo(dst, x, k []byte) {
	var t [16]byte
	xorBlock(t[:], x, k)
	c.sl1(t[:])
	c.diffusionA(dst, t[:])
}

// fe is the even round function
func (c *ariaCipher) fe(dst, x, k []byte) {
	var t [16]byte
	xorBlock(t[:], x, k)
	c.sl2(t[:])
	c.diffusionA(dst, t[:])
}

// sl1 applies substitution layer type 1 (odd rounds)
func (c *ariaCipher) sl1(x []byte) {
	x[0] = sbox1[x[0]]
	x[1] = sbox2[x[1]]
	x[2] = sbox1Inv[x[2]]
	x[3] = sbox2Inv[x[3]]
	x[4] = sbox1[x[4]]
	x[5] = sbox2[x[5]]
	x[6] = sbox1Inv[x[6]]
	x[7] = sbox2Inv[x[7]]
	x[8] = sbox1[x[8]]
	x[9] = sbox2[x[9]]
	x[10] = sbox1Inv[x[10]]
	x[11] = sbox2Inv[x[11]]
	x[12] = sbox1[x[12]]
	x[13] = sbox2[x[13]]
	x[14] = sbox1Inv[x[14]]
	x[15] = sbox2Inv[x[15]]
}

// sl2 applies substitution layer type 2 (even rounds)
func (c *ariaCipher) sl2(x []byte) {
	x[0] = sbox1Inv[x[0]]
	x[1] = sbox2Inv[x[1]]
	x[2] = sbox1[x[2]]
	x[3] = sbox2[x[3]]
	x[4] = sbox1Inv[x[4]]
	x[5] = sbox2Inv[x[5]]
	x[6] = sbox1[x[6]]
	x[7] = sbox2[x[7]]
	x[8] = sbox1Inv[x[8]]
	x[9] = sbox2Inv[x[9]]
	x[10] = sbox1[x[10]]
	x[11] = sbox2[x[11]]
	x[12] = sbox1Inv[x[12]]
	x[13] = sbox2Inv[x[13]]
	x[14] = sbox1[x[14]]
	x[15] = sbox2[x[15]]
}

// diffusionA applies the diffusion layer
func (c *ariaCipher) diffusionA(dst, x []byte) {
	dst[0] = x[3] ^ x[4] ^ x[6] ^ x[8] ^ x[9] ^ x[13] ^ x[14]
	dst[1] = x[2] ^ x[5] ^ x[7] ^ x[8] ^ x[9] ^ x[12] ^ x[15]
	dst[2] = x[1] ^ x[4] ^ x[6] ^ x[10] ^ x[11] ^ x[12] ^ x[15]
	dst[3] = x[0] ^ x[5] ^ x[7] ^ x[10] ^ x[11] ^ x[13] ^ x[14]
	dst[4] = x[0] ^ x[2] ^ x[5] ^ x[8] ^ x[11] ^ x[14] ^ x[15]
	dst[5] = x[1] ^ x[3] ^ x[4] ^ x[9] ^ x[10] ^ x[14] ^ x[15]
	dst[6] = x[0] ^ x[2] ^ x[7] ^ x[9] ^ x[10] ^ x[12] ^ x[13]
	dst[7] = x[1] ^ x[3] ^ x[6] ^ x[8] ^ x[11] ^ x[12] ^ x[13]
	dst[8] = x[0] ^ x[1] ^ x[4] ^ x[7] ^ x[10] ^ x[13] ^ x[15]
	dst[9] = x[0] ^ x[1] ^ x[5] ^ x[6] ^ x[11] ^ x[12] ^ x[14]
	dst[10] = x[2] ^ x[3] ^ x[5] ^ x[6] ^ x[8] ^ x[13] ^ x[15]
	dst[11] = x[2] ^ x[3] ^ x[4] ^ x[7] ^ x[9] ^ x[12] ^ x[14]
	dst[12] = x[1] ^ x[2] ^ x[6] ^ x[7] ^ x[9] ^ x[11] ^ x[12]
	dst[13] = x[0] ^ x[3] ^ x[6] ^ x[7] ^ x[8] ^ x[10] ^ x[13]
	dst[14] = x[0] ^ x[3] ^ x[4] ^ x[5] ^ x[9] ^ x[11] ^ x[14]
	dst[15] = x[1] ^ x[2] ^ x[4] ^ x[5] ^ x[8] ^ x[10] ^ x[15]
}

// encrypt performs ARIA encryption
func (c *ariaCipher) encrypt(dst, src []byte) {
	var state [16]byte
	copy(state[:], src)

	// Rounds 1 to rounds-1
	for i := 0; i < c.rounds-1; i++ {
		roundKey := c.ek[i*16 : (i+1)*16]
		xorBlock(state[:], state[:], roundKey)

		if i%2 == 0 {
			c.sl1(state[:])
		} else {
			c.sl2(state[:])
		}

		var temp [16]byte
		c.diffusionA(temp[:], state[:])
		copy(state[:], temp[:])
	}

	// Last round (no diffusion)
	roundKey := c.ek[(c.rounds-1)*16 : c.rounds*16]
	xorBlock(state[:], state[:], roundKey)

	if (c.rounds-1)%2 == 0 {
		c.sl1(state[:])
	} else {
		c.sl2(state[:])
	}

	// Final key addition
	finalKey := c.ek[c.rounds*16 : (c.rounds+1)*16]
	xorBlock(dst, state[:], finalKey)
}

// decrypt performs ARIA decryption
func (c *ariaCipher) decrypt(dst, src []byte) {
	var state [16]byte
	copy(state[:], src)

	// Rounds 1 to rounds-1
	for i := 0; i < c.rounds-1; i++ {
		roundKey := c.dk[i*16 : (i+1)*16]
		xorBlock(state[:], state[:], roundKey)

		if i%2 == 0 {
			c.sl1(state[:])
		} else {
			c.sl2(state[:])
		}

		var temp [16]byte
		c.diffusionA(temp[:], state[:])
		copy(state[:], temp[:])
	}

	// Last round (no diffusion)
	roundKey := c.dk[(c.rounds-1)*16 : c.rounds*16]
	xorBlock(state[:], state[:], roundKey)

	if (c.rounds-1)%2 == 0 {
		c.sl1(state[:])
	} else {
		c.sl2(state[:])
	}

	// Final key addition
	finalKey := c.dk[c.rounds*16 : (c.rounds+1)*16]
	xorBlock(dst, state[:], finalKey)
}
