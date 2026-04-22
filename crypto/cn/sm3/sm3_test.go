package sm3

import (
	"encoding/hex"
	"testing"
)

// Test vectors from GB/T 32905-2016 Appendix A
func TestSM3(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "Example 1: abc",
			input:  "abc",
			expect: "66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0",
		},
		{
			name:   "Example 2: 64 bytes repeated",
			input:  "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd",
			expect: "debe9ff92275b8a138604889c18e5a4d6fdb70e5387e5765293dcba39c0c5732",
		},
		{
			name:   "Empty string",
			input:  "",
			expect: "1ab21d8355cfa17f8e61194831e81a8f22bec8c728fefb747ed035eb5082aa2b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sum([]byte(tt.input))
			got := hex.EncodeToString(result[:])
			if got != tt.expect {
				t.Errorf("SM3(%q) = %s, want %s", tt.input, got, tt.expect)
			}
		})
	}
}

func TestSM3Incremental(t *testing.T) {
	// Test incremental hashing produces same result
	input := "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd"
	expected := Sum([]byte(input))

	h := New()
	h.Write([]byte("abcdabcd"))
	h.Write([]byte("abcdabcd"))
	h.Write([]byte("abcdabcd"))
	h.Write([]byte("abcdabcd"))
	h.Write([]byte("abcdabcd"))
	h.Write([]byte("abcdabcd"))
	h.Write([]byte("abcdabcd"))
	h.Write([]byte("abcdabcd"))

	var result [Size]byte
	copy(result[:], h.Sum(nil))

	if result != expected {
		t.Errorf("Incremental hash mismatch")
	}
}

func TestSM3Size(t *testing.T) {
	h := New()
	if h.Size() != Size {
		t.Errorf("Size() = %d, want %d", h.Size(), Size)
	}
	if h.BlockSize() != BlockSize {
		t.Errorf("BlockSize() = %d, want %d", h.BlockSize(), BlockSize)
	}
}

func BenchmarkSM3(b *testing.B) {
	data := make([]byte, 1024)
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum(data)
	}
}
