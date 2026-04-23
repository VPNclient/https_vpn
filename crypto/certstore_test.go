package crypto_test

import (
	"reflect"
	"testing"

	"github.com/nativemind/https-vpn/crypto"
	// Import providers to register them
	_ "github.com/nativemind/https-vpn/crypto/cn"
	_ "github.com/nativemind/https-vpn/crypto/us"
)

func TestParseProviderPriority(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single provider",
			input:    "us",
			expected: []string{"us"},
		},
		{
			name:     "multiple providers",
			input:    "cn,us",
			expected: []string{"cn", "us"},
		},
		{
			name:     "with spaces",
			input:    " cn , us ",
			expected: []string{"cn", "us"},
		},
		{
			name:     "mixed with cipher suite names",
			input:    "cn,TLS_AES_128_GCM_SHA256,us",
			expected: []string{"cn", "us"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{"us"},
		},
		{
			name:     "no valid providers",
			input:    "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384",
			expected: []string{"us"},
		},
		{
			name:     "duplicates removed",
			input:    "cn,cn,us,us",
			expected: []string{"cn", "us"},
		},
		{
			name:     "case insensitive",
			input:    "CN,US",
			expected: []string{"cn", "us"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crypto.ParseProviderPriority(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseProviderPriority(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsSM4Suite(t *testing.T) {
	tests := []struct {
		suite    uint16
		expected bool
	}{
		{0x00C6, true},  // TLS_SM4_GCM_SM3
		{0x00C7, true},  // TLS_SM4_CCM_SM3
		{0x1301, false}, // TLS_AES_128_GCM_SHA256
		{0x1302, false}, // TLS_AES_256_GCM_SHA384
		{0x0000, false},
		{0xFFFF, false},
	}

	for _, tt := range tests {
		result := crypto.IsSM4Suite(tt.suite)
		if result != tt.expected {
			t.Errorf("IsSM4Suite(0x%04X) = %v, expected %v", tt.suite, result, tt.expected)
		}
	}
}

func TestIsGOSTSuite(t *testing.T) {
	tests := []struct {
		suite    uint16
		expected bool
	}{
		{0xFF85, true}, // GOST suite
		{0xFF86, true}, // GOST suite
		{0xFF87, true}, // GOST suite
		{0xFF88, true}, // GOST suite
		{0xFF84, false},
		{0xFF89, false},
		{0x1301, false}, // TLS_AES_128_GCM_SHA256
		{0x00C6, false}, // SM4 suite
	}

	for _, tt := range tests {
		result := crypto.IsGOSTSuite(tt.suite)
		if result != tt.expected {
			t.Errorf("IsGOSTSuite(0x%04X) = %v, expected %v", tt.suite, result, tt.expected)
		}
	}
}
