// Package kalyna реалізує блочний шифр Калина за стандартом ДСТУ 7624:2014.
// Підтримує варіанти Калина-128/128, Калина-128/256, Калина-256/256,
// Калина-256/512 та Калина-512/512.
package kalyna

import (
	"crypto/cipher"
	"encoding/binary"
	"errors"
)

// Розміри блоків та ключів у байтах
const (
	BlockSize128 = 16 // 128 біт
	BlockSize256 = 32 // 256 біт
	BlockSize512 = 64 // 512 біт

	KeySize128 = 16 // 128 біт
	KeySize256 = 32 // 256 біт
	KeySize512 = 64 // 512 біт
)

// Кількість раундів для різних варіантів
const (
	rounds128_128 = 10
	rounds128_256 = 14
	rounds256_256 = 14
	rounds256_512 = 18
	rounds512_512 = 18
)

// Cipher реалізує cipher.Block для шифру Калина
type Cipher struct {
	blockSize int
	keySize   int
	rounds    int
	roundKeys [][]uint64 // Раундові ключі
}

// NewCipher128 створює новий шифр Калина-128 з 128-бітним ключем
func NewCipher128(key []byte) (cipher.Block, error) {
	if len(key) != KeySize128 {
		return nil, errors.New("kalyna: недійсний розмір ключа, потрібно 16 байт")
	}
	return newCipher(key, BlockSize128, rounds128_128)
}

// NewCipher256 створює новий шифр Калина-256 з 256-бітним ключем
func NewCipher256(key []byte) (cipher.Block, error) {
	if len(key) != KeySize256 {
		return nil, errors.New("kalyna: недійсний розмір ключа, потрібно 32 байти")
	}
	return newCipher(key, BlockSize256, rounds256_256)
}

// NewCipher512 створює новий шифр Калина-512 з 512-бітним ключем
func NewCipher512(key []byte) (cipher.Block, error) {
	if len(key) != KeySize512 {
		return nil, errors.New("kalyna: недійсний розмір ключа, потрібно 64 байти")
	}
	return newCipher(key, BlockSize512, rounds512_512)
}

// NewCipher створює шифр Калина з автовизначенням параметрів за розміром ключа
func NewCipher(key []byte) (cipher.Block, error) {
	switch len(key) {
	case KeySize128:
		return NewCipher128(key)
	case KeySize256:
		return NewCipher256(key)
	case KeySize512:
		return NewCipher512(key)
	default:
		return nil, errors.New("kalyna: недійсний розмір ключа")
	}
}

func newCipher(key []byte, blockSize, rounds int) (*Cipher, error) {
	c := &Cipher{
		blockSize: blockSize,
		keySize:   len(key),
		rounds:    rounds,
	}

	c.expandKey(key)
	return c, nil
}

// BlockSize повертає розмір блоку в байтах
func (c *Cipher) BlockSize() int {
	return c.blockSize
}

// Encrypt шифрує один блок
func (c *Cipher) Encrypt(dst, src []byte) {
	if len(src) < c.blockSize || len(dst) < c.blockSize {
		panic("kalyna: вхідний або вихідний буфер занадто малий")
	}

	state := c.bytesToState(src)
	state = c.encryptBlock(state)
	c.stateToBytes(state, dst)
}

// Decrypt розшифровує один блок
func (c *Cipher) Decrypt(dst, src []byte) {
	if len(src) < c.blockSize || len(dst) < c.blockSize {
		panic("kalyna: вхідний або вихідний буфер занадто малий")
	}

	state := c.bytesToState(src)
	state = c.decryptBlock(state)
	c.stateToBytes(state, dst)
}

// encryptBlock шифрує один блок у вигляді стану
func (c *Cipher) encryptBlock(state []uint64) []uint64 {
	cols := c.blockSize / 8

	// Початкове додавання ключа
	state = addRoundKey(state, c.roundKeys[0])

	// Основні раунди
	for r := 1; r < c.rounds; r++ {
		state = subBytes(state, cols)
		state = shiftRows(state, cols)
		state = mixColumns(state, cols)
		state = xorRoundKey(state, c.roundKeys[r])
	}

	// Останній раунд
	state = subBytes(state, cols)
	state = shiftRows(state, cols)
	state = mixColumns(state, cols)
	state = addRoundKey(state, c.roundKeys[c.rounds])

	return state
}

// decryptBlock розшифровує один блок
func (c *Cipher) decryptBlock(state []uint64) []uint64 {
	cols := c.blockSize / 8

	// Початкове віднімання ключа
	state = subRoundKey(state, c.roundKeys[c.rounds])
	state = invMixColumns(state, cols)
	state = invShiftRows(state, cols)
	state = invSubBytes(state, cols)

	// Основні раунди
	for r := c.rounds - 1; r > 0; r-- {
		state = xorRoundKey(state, c.roundKeys[r])
		state = invMixColumns(state, cols)
		state = invShiftRows(state, cols)
		state = invSubBytes(state, cols)
	}

	// Останнє віднімання ключа
	state = subRoundKey(state, c.roundKeys[0])

	return state
}

// bytesToState конвертує байти у стан
func (c *Cipher) bytesToState(b []byte) []uint64 {
	cols := c.blockSize / 8
	state := make([]uint64, cols)
	for i := 0; i < cols; i++ {
		state[i] = binary.LittleEndian.Uint64(b[i*8:])
	}
	return state
}

// stateToBytes конвертує стан у байти
func (c *Cipher) stateToBytes(state []uint64, b []byte) {
	for i := 0; i < len(state); i++ {
		binary.LittleEndian.PutUint64(b[i*8:], state[i])
	}
}

// expandKey розгортає ключ у раундові ключі
func (c *Cipher) expandKey(key []byte) {
	cols := c.blockSize / 8
	c.roundKeys = make([][]uint64, c.rounds+1)

	// Ініціалізація
	kt := make([]uint64, cols)
	k := make([]uint64, c.keySize/8)

	for i := 0; i < len(k); i++ {
		k[i] = binary.LittleEndian.Uint64(key[i*8:])
	}

	// Обчислення проміжного ключа KT
	kt = c.computeKT(k, cols)

	// Генерація раундових ключів
	for i := 0; i <= c.rounds; i++ {
		c.roundKeys[i] = make([]uint64, cols)
		c.computeRoundKey(i, kt, k, cols)
	}
}

// computeKT обчислює проміжний ключ
func (c *Cipher) computeKT(k []uint64, cols int) []uint64 {
	kt := make([]uint64, cols)

	// Перша частина ключа
	k0 := k[:cols]

	// Застосовуємо раундову функцію
	state := make([]uint64, cols)
	copy(state, k0)

	state = addRoundKeyConst(state, cols, 0)
	state = subBytes(state, cols)
	state = shiftRows(state, cols)
	state = mixColumns(state, cols)
	state = xorState(state, k0)
	state = addRoundKeyConst(state, cols, 1)
	state = subBytes(state, cols)
	state = shiftRows(state, cols)
	state = mixColumns(state, cols)

	copy(kt, state)
	return kt
}

// computeRoundKey обчислює раундовий ключ
func (c *Cipher) computeRoundKey(round int, kt, k []uint64, cols int) {
	// Константа для even/odd раундів
	var tmv uint64
	if round%2 == 0 {
		tmv = 0x0001000100010001
	} else {
		tmv = 0x0001000100010001
	}

	// Зсув та маскування
	shift := round * (c.blockSize / 4)
	shift %= c.blockSize * 8

	state := make([]uint64, cols)
	copy(state, kt)

	// Додаємо константу
	for i := 0; i < cols; i++ {
		state[i] += tmv << ((uint(round) * 8) % 64)
	}

	// Застосовуємо перетворення
	state = subBytes(state, cols)
	state = shiftRows(state, cols)
	state = mixColumns(state, cols)

	// XOR з частиною ключа
	kpart := c.keySize / 8
	offset := (round * cols / 2) % kpart
	for i := 0; i < cols; i++ {
		state[i] ^= k[(offset+i)%kpart]
	}

	// Зсув раундового ключа
	state = rotateLeft(state, shift%512)

	copy(c.roundKeys[round], state)
}

// addRoundKey додає раундовий ключ (mod 2^64)
func addRoundKey(state, key []uint64) []uint64 {
	result := make([]uint64, len(state))
	for i := range state {
		result[i] = state[i] + key[i]
	}
	return result
}

// subRoundKey віднімає раундовий ключ (mod 2^64)
func subRoundKey(state, key []uint64) []uint64 {
	result := make([]uint64, len(state))
	for i := range state {
		result[i] = state[i] - key[i]
	}
	return result
}

// xorRoundKey XOR з раундовим ключем
func xorRoundKey(state, key []uint64) []uint64 {
	result := make([]uint64, len(state))
	for i := range state {
		result[i] = state[i] ^ key[i]
	}
	return result
}

// xorState XOR двох станів
func xorState(a, b []uint64) []uint64 {
	result := make([]uint64, len(a))
	for i := range a {
		result[i] = a[i] ^ b[i]
	}
	return result
}

// addRoundKeyConst додає константу раунду
func addRoundKeyConst(state []uint64, cols int, round int) []uint64 {
	result := make([]uint64, cols)
	copy(result, state)
	for i := 0; i < cols; i++ {
		result[i] += uint64((cols-i)<<4) | uint64(round)
	}
	return result
}

// subBytes застосовує S-box
func subBytes(state []uint64, cols int) []uint64 {
	result := make([]uint64, cols)
	for i := 0; i < cols; i++ {
		var val uint64
		for j := 0; j < 8; j++ {
			b := byte(state[i] >> (j * 8))
			sb := sbox[j%4][b]
			val |= uint64(sb) << (j * 8)
		}
		result[i] = val
	}
	return result
}

// invSubBytes застосовує інверсний S-box
func invSubBytes(state []uint64, cols int) []uint64 {
	result := make([]uint64, cols)
	for i := 0; i < cols; i++ {
		var val uint64
		for j := 0; j < 8; j++ {
			b := byte(state[i] >> (j * 8))
			sb := invSbox[j%4][b]
			val |= uint64(sb) << (j * 8)
		}
		result[i] = val
	}
	return result
}

// shiftRows циклічний зсув рядків
func shiftRows(state []uint64, cols int) []uint64 {
	result := make([]uint64, cols)

	var shifts [8]int
	switch cols {
	case 2: // 128 біт
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	case 4: // 256 біт
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	case 8: // 512 біт
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 11}
	default:
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	}

	for row := 0; row < 8; row++ {
		shift := shifts[row] % cols
		for col := 0; col < cols; col++ {
			srcCol := (col + shift) % cols
			b := byte(state[srcCol] >> (row * 8))
			result[col] |= uint64(b) << (row * 8)
		}
	}

	return result
}

// invShiftRows інверсний циклічний зсув рядків
func invShiftRows(state []uint64, cols int) []uint64 {
	result := make([]uint64, cols)

	var shifts [8]int
	switch cols {
	case 2:
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	case 4:
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	case 8:
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 11}
	default:
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	}

	for row := 0; row < 8; row++ {
		shift := shifts[row] % cols
		for col := 0; col < cols; col++ {
			srcCol := (col + cols - shift) % cols
			b := byte(state[srcCol] >> (row * 8))
			result[col] |= uint64(b) << (row * 8)
		}
	}

	return result
}

// mixColumns множення на MDS матрицю
func mixColumns(state []uint64, cols int) []uint64 {
	result := make([]uint64, cols)

	for col := 0; col < cols; col++ {
		var column [8]byte
		for row := 0; row < 8; row++ {
			column[row] = byte(state[col] >> (row * 8))
		}

		var newColumn [8]byte
		for row := 0; row < 8; row++ {
			var sum byte
			for k := 0; k < 8; k++ {
				sum ^= gfMul(mdsMatrix[row][k], column[k])
			}
			newColumn[row] = sum
		}

		for row := 0; row < 8; row++ {
			result[col] |= uint64(newColumn[row]) << (row * 8)
		}
	}

	return result
}

// invMixColumns інверсне множення на MDS матрицю
func invMixColumns(state []uint64, cols int) []uint64 {
	result := make([]uint64, cols)

	for col := 0; col < cols; col++ {
		var column [8]byte
		for row := 0; row < 8; row++ {
			column[row] = byte(state[col] >> (row * 8))
		}

		var newColumn [8]byte
		for row := 0; row < 8; row++ {
			var sum byte
			for k := 0; k < 8; k++ {
				sum ^= gfMul(invMdsMatrix[row][k], column[k])
			}
			newColumn[row] = sum
		}

		for row := 0; row < 8; row++ {
			result[col] |= uint64(newColumn[row]) << (row * 8)
		}
	}

	return result
}

// gfMul множення в GF(2^8) з поліномом x^8 + x^4 + x^3 + x + 1 (0x11B для Калини)
func gfMul(a, b byte) byte {
	var result byte
	for i := 0; i < 8; i++ {
		if b&1 != 0 {
			result ^= a
		}
		hi := a & 0x80
		a <<= 1
		if hi != 0 {
			a ^= 0x1D // Редукційний поліном
		}
		b >>= 1
	}
	return result
}

// rotateLeft циклічний зсув стану вліво на bits біт
func rotateLeft(state []uint64, bits int) []uint64 {
	if bits == 0 {
		result := make([]uint64, len(state))
		copy(result, state)
		return result
	}

	totalBits := len(state) * 64
	bits = bits % totalBits

	// Конвертуємо в байти для простоти
	bytes := make([]byte, len(state)*8)
	for i, v := range state {
		binary.LittleEndian.PutUint64(bytes[i*8:], v)
	}

	// Зсув на байти
	byteShift := bits / 8
	bitShift := bits % 8

	result := make([]byte, len(bytes))
	for i := 0; i < len(bytes); i++ {
		srcIdx := (i + byteShift) % len(bytes)
		nextIdx := (srcIdx + 1) % len(bytes)
		result[i] = (bytes[srcIdx] << bitShift) | (bytes[nextIdx] >> (8 - bitShift))
	}

	// Конвертуємо назад
	stateResult := make([]uint64, len(state))
	for i := range stateResult {
		stateResult[i] = binary.LittleEndian.Uint64(result[i*8:])
	}

	return stateResult
}
