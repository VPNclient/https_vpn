// Package kupyna реалізує хеш-функцію Купина за стандартом ДСТУ 7564:2014.
// Підтримує розміри хешу 256 та 512 біт.
package kupyna

import (
	"encoding/binary"
	"hash"
)

// Розміри хешу в байтах
const (
	Size256   = 32 // 256 біт
	Size512   = 64 // 512 біт
	BlockSize = 64 // Розмір блоку в байтах (512 біт)
)

// Кількість раундів для різних розмірів
const (
	rounds256 = 10
	rounds512 = 14
)

// Кількість стовпців у матриці стану
const (
	cols256 = 8  // 8 стовпців для 256/512 біт входу
	cols512 = 16 // 16 стовпців для 1024 біт входу (внутрішній стан 512)
)

// Digest представляє стан хеш-функції Купина
type Digest struct {
	state  [cols512]uint64 // Внутрішній стан
	buf    [BlockSize * 2]byte
	bufLen int
	msgLen uint64
	size   int // Розмір хешу (32 або 64 байти)
	cols   int // Кількість стовпців (8 або 16)
	rounds int // Кількість раундів
}

// New256 створює новий хеш Купина-256
func New256() hash.Hash {
	d := &Digest{
		size:   Size256,
		cols:   cols256,
		rounds: rounds256,
	}
	d.Reset()
	return d
}

// New512 створює новий хеш Купина-512
func New512() hash.Hash {
	d := &Digest{
		size:   Size512,
		cols:   cols512,
		rounds: rounds512,
	}
	d.Reset()
	return d
}

// New створює новий хеш Купина-512 (за замовчуванням)
func New() hash.Hash {
	return New512()
}

// Reset скидає стан хешу до початкового
func (d *Digest) Reset() {
	for i := range d.state {
		d.state[i] = 0
	}
	// Ініціалізація IV: перший байт = розмір хешу в бітах
	d.state[0] = uint64(d.size) << 3
	d.bufLen = 0
	d.msgLen = 0
}

// Size повертає розмір хешу в байтах
func (d *Digest) Size() int {
	return d.size
}

// BlockSize повертає розмір блоку в байтах
func (d *Digest) BlockSize() int {
	if d.cols == cols512 {
		return BlockSize * 2
	}
	return BlockSize
}

// Write додає дані до хешу
func (d *Digest) Write(p []byte) (n int, err error) {
	n = len(p)
	d.msgLen += uint64(n)

	blockSize := d.BlockSize()

	// Якщо є дані в буфері, спробуємо заповнити блок
	if d.bufLen > 0 {
		need := blockSize - d.bufLen
		if len(p) < need {
			copy(d.buf[d.bufLen:], p)
			d.bufLen += len(p)
			return n, nil
		}
		copy(d.buf[d.bufLen:], p[:need])
		d.processBlock(d.buf[:blockSize])
		p = p[need:]
		d.bufLen = 0
	}

	// Обробляємо повні блоки
	for len(p) >= blockSize {
		d.processBlock(p[:blockSize])
		p = p[blockSize:]
	}

	// Зберігаємо залишок
	if len(p) > 0 {
		copy(d.buf[:], p)
		d.bufLen = len(p)
	}

	return n, nil
}

// Sum завершує обчислення хешу і повертає результат
func (d *Digest) Sum(in []byte) []byte {
	// Копіюємо стан для можливості продовження
	d0 := *d
	hash := d0.finalize()
	return append(in, hash...)
}

// finalize завершує обчислення хешу
func (d *Digest) finalize() []byte {
	blockSize := d.BlockSize()

	// Додаємо padding
	padLen := blockSize - 12 - d.bufLen
	if padLen <= 0 {
		padLen += blockSize
	}

	// Padding: 0x80 || 0x00... || length (96 біт)
	padding := make([]byte, padLen+12)
	padding[0] = 0x80
	// Довжина повідомлення в бітах (96 біт = 12 байт)
	msgBits := d.msgLen << 3
	binary.LittleEndian.PutUint64(padding[padLen:], msgBits)
	binary.LittleEndian.PutUint32(padding[padLen+8:], 0)

	d.Write(padding)

	// Фінальне перетворення
	d.outputTransform()

	// Вибираємо праву половину стану
	result := make([]byte, d.size)
	offset := (d.cols - d.size/8) * 8
	for i := 0; i < d.size/8; i++ {
		binary.LittleEndian.PutUint64(result[i*8:], d.state[offset/8+i])
	}

	return result
}

// processBlock обробляє один блок даних
func (d *Digest) processBlock(block []byte) {
	var m [cols512]uint64

	// Завантажуємо блок у little-endian
	for i := 0; i < d.cols; i++ {
		m[i] = binary.LittleEndian.Uint64(block[i*8:])
	}

	// Обчислюємо T_xor(M, H)
	var t [cols512]uint64
	for i := 0; i < d.cols; i++ {
		t[i] = m[i] ^ d.state[i]
	}

	// Застосовуємо раундову функцію P
	pState := d.roundP(t[:d.cols])

	// Застосовуємо раундову функцію Q до M
	qState := d.roundQ(m[:d.cols])

	// XOR результатів
	for i := 0; i < d.cols; i++ {
		d.state[i] ^= pState[i] ^ qState[i]
	}
}

// outputTransform виконує фінальне перетворення
func (d *Digest) outputTransform() {
	// P(H) XOR H
	pState := d.roundP(d.state[:d.cols])
	for i := 0; i < d.cols; i++ {
		d.state[i] ^= pState[i]
	}
}

// roundP застосовує раундову функцію P
func (d *Digest) roundP(state []uint64) []uint64 {
	result := make([]uint64, d.cols)
	copy(result, state)

	for r := 0; r < d.rounds; r++ {
		// AddRoundConstant для P
		for i := 0; i < d.cols; i++ {
			result[i] ^= uint64(i<<4) ^ uint64(r)
		}
		result = d.subBytes(result)
		result = d.shiftBytes(result)
		result = d.mixColumns(result)
	}

	return result
}

// roundQ застосовує раундову функцію Q
func (d *Digest) roundQ(state []uint64) []uint64 {
	result := make([]uint64, d.cols)
	copy(result, state)

	for r := 0; r < d.rounds; r++ {
		// AddRoundConstant для Q
		for i := 0; i < d.cols; i++ {
			// Q використовує інші константи
			result[i] += uint64((d.cols-1-i)<<4) ^ uint64((d.rounds-1-r)<<56)
		}
		result = d.subBytes(result)
		result = d.shiftBytes(result)
		result = d.mixColumns(result)
	}

	return result
}

// subBytes застосовує S-box до кожного байту
func (d *Digest) subBytes(state []uint64) []uint64 {
	result := make([]uint64, d.cols)
	for i := 0; i < d.cols; i++ {
		var val uint64
		for j := 0; j < 8; j++ {
			b := byte(state[i] >> (j * 8))
			// Використовуємо різні S-box для різних рядків
			sb := sbox[j%4][b]
			val |= uint64(sb) << (j * 8)
		}
		result[i] = val
	}
	return result
}

// shiftBytes виконує циклічний зсув рядків
func (d *Digest) shiftBytes(state []uint64) []uint64 {
	result := make([]uint64, d.cols)

	// Визначаємо зсуви для кожного рядка
	var shifts [8]int
	if d.cols == cols256 {
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	} else {
		shifts = [8]int{0, 1, 2, 3, 4, 5, 6, 11}
	}

	for row := 0; row < 8; row++ {
		shift := shifts[row]
		for col := 0; col < d.cols; col++ {
			srcCol := (col + shift) % d.cols
			b := byte(state[srcCol] >> (row * 8))
			result[col] |= uint64(b) << (row * 8)
		}
	}

	return result
}

// mixColumns виконує множення на MDS матрицю
func (d *Digest) mixColumns(state []uint64) []uint64 {
	result := make([]uint64, d.cols)

	for col := 0; col < d.cols; col++ {
		var column [8]byte
		for row := 0; row < 8; row++ {
			column[row] = byte(state[col] >> (row * 8))
		}

		// Множення на MDS матрицю в GF(2^8)
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

// gfMul множення в GF(2^8) з поліномом x^8 + x^4 + x^3 + x^2 + 1 (0x11D)
func gfMul(a, b byte) byte {
	var result byte
	for i := 0; i < 8; i++ {
		if b&1 != 0 {
			result ^= a
		}
		hi := a & 0x80
		a <<= 1
		if hi != 0 {
			a ^= 0x1D // x^8 + x^4 + x^3 + x^2 + 1 mod x^8
		}
		b >>= 1
	}
	return result
}

// Sum512 обчислює хеш Купина-512 від даних
func Sum512(data []byte) [Size512]byte {
	h := New512()
	h.Write(data)
	var result [Size512]byte
	copy(result[:], h.Sum(nil))
	return result
}

// Sum256 обчислює хеш Купина-256 від даних
func Sum256(data []byte) [Size256]byte {
	h := New256()
	h.Write(data)
	var result [Size256]byte
	copy(result[:], h.Sum(nil))
	return result
}
