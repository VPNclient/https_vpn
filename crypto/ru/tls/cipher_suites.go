// GOST TLS cipher suites per RFC 9189.
package tls

// GOST TLS 1.2 cipher suites
const (
	// With Kuznyechik (Grasshopper)
	TLS_GOSTR341112_256_WITH_KUZNYECHIK_CTR_OMAC uint16 = 0xC100
	TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_L    uint16 = 0xC101
	TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_S    uint16 = 0xC102

	// With Magma
	TLS_GOSTR341112_256_WITH_MAGMA_CTR_OMAC uint16 = 0xC103
	TLS_GOSTR341112_256_WITH_MAGMA_MGM_L    uint16 = 0xC104
	TLS_GOSTR341112_256_WITH_MAGMA_MGM_S    uint16 = 0xC105

	// 512-bit signature variants
	TLS_GOSTR341112_512_WITH_KUZNYECHIK_CTR_OMAC uint16 = 0xC106
	TLS_GOSTR341112_512_WITH_KUZNYECHIK_MGM_L    uint16 = 0xC107
	TLS_GOSTR341112_512_WITH_KUZNYECHIK_MGM_S    uint16 = 0xC108

	TLS_GOSTR341112_512_WITH_MAGMA_CTR_OMAC uint16 = 0xC109
	TLS_GOSTR341112_512_WITH_MAGMA_MGM_L    uint16 = 0xC10A
	TLS_GOSTR341112_512_WITH_MAGMA_MGM_S    uint16 = 0xC10B
)

// GOST signature algorithms for TLS
const (
	SignatureGOSTR34102012_256 uint16 = 0x0709 // GOST R 34.10-2012 with 256-bit key
	SignatureGOSTR34102012_512 uint16 = 0x070A // GOST R 34.10-2012 with 512-bit key
)

// GOST named groups (curves) for TLS key exchange
const (
	CurveGOSTR34102012_256_A uint16 = 0x0022 // id-tc26-gost-3410-2012-256-paramSetA
	CurveGOSTR34102012_256_B uint16 = 0x0023 // id-tc26-gost-3410-2012-256-paramSetB
	CurveGOSTR34102012_256_C uint16 = 0x0024 // id-tc26-gost-3410-2012-256-paramSetC
	CurveGOSTR34102012_256_D uint16 = 0x0025 // id-tc26-gost-3410-2012-256-paramSetD

	CurveGOSTR34102012_512_A uint16 = 0x0026 // id-tc26-gost-3410-2012-512-paramSetA
	CurveGOSTR34102012_512_B uint16 = 0x0027 // id-tc26-gost-3410-2012-512-paramSetB
	CurveGOSTR34102012_512_C uint16 = 0x0028 // id-tc26-gost-3410-2012-512-paramSetC
)

// TLS protocol versions
const (
	VersionTLS12 uint16 = 0x0303
	VersionTLS13 uint16 = 0x0304
)

// TLS record types
const (
	RecordTypeChangeCipherSpec uint8 = 20
	RecordTypeAlert            uint8 = 21
	RecordTypeHandshake        uint8 = 22
	RecordTypeApplicationData  uint8 = 23
)

// TLS handshake types
const (
	HandshakeTypeClientHello        uint8 = 1
	HandshakeTypeServerHello        uint8 = 2
	HandshakeTypeNewSessionTicket   uint8 = 4
	HandshakeTypeEndOfEarlyData     uint8 = 5
	HandshakeTypeEncryptedExtensions uint8 = 8
	HandshakeTypeCertificate        uint8 = 11
	HandshakeTypeServerKeyExchange  uint8 = 12
	HandshakeTypeCertificateRequest uint8 = 13
	HandshakeTypeServerHelloDone    uint8 = 14
	HandshakeTypeCertificateVerify  uint8 = 15
	HandshakeTypeClientKeyExchange  uint8 = 16
	HandshakeTypeFinished           uint8 = 20
)

// TLS extension types
const (
	ExtensionServerName          uint16 = 0
	ExtensionSupportedGroups     uint16 = 10
	ExtensionSignatureAlgorithms uint16 = 13
	ExtensionALPN                uint16 = 16
	ExtensionSupportedVersions   uint16 = 43
	ExtensionKeyShare            uint16 = 51
)

// CipherSuiteInfo contains metadata about a cipher suite.
type CipherSuiteInfo struct {
	ID           uint16
	Name         string
	KeySize      int  // Key size in bytes
	IVSize       int  // IV/nonce size in bytes
	TagSize      int  // Authentication tag size in bytes
	Hash         int  // Hash algorithm (256 or 512)
	Cipher       int  // Cipher type (Kuznyechik or Magma)
	Mode         int  // Mode (CTR_OMAC or MGM)
}

// Cipher types
const (
	CipherKuznyechik = iota
	CipherMagma
)

// Mode types
const (
	ModeCTR_OMAC = iota
	ModeMGM_L
	ModeMGM_S
)

// CipherSuites maps cipher suite IDs to their info.
var CipherSuites = map[uint16]CipherSuiteInfo{
	TLS_GOSTR341112_256_WITH_KUZNYECHIK_CTR_OMAC: {
		ID: TLS_GOSTR341112_256_WITH_KUZNYECHIK_CTR_OMAC,
		Name: "TLS_GOSTR341112_256_WITH_KUZNYECHIK_CTR_OMAC",
		KeySize: 32, IVSize: 8, TagSize: 16,
		Hash: 256, Cipher: CipherKuznyechik, Mode: ModeCTR_OMAC,
	},
	TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_L: {
		ID: TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_L,
		Name: "TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_L",
		KeySize: 32, IVSize: 16, TagSize: 16,
		Hash: 256, Cipher: CipherKuznyechik, Mode: ModeMGM_L,
	},
	TLS_GOSTR341112_256_WITH_MAGMA_CTR_OMAC: {
		ID: TLS_GOSTR341112_256_WITH_MAGMA_CTR_OMAC,
		Name: "TLS_GOSTR341112_256_WITH_MAGMA_CTR_OMAC",
		KeySize: 32, IVSize: 4, TagSize: 8,
		Hash: 256, Cipher: CipherMagma, Mode: ModeCTR_OMAC,
	},
	TLS_GOSTR341112_256_WITH_MAGMA_MGM_L: {
		ID: TLS_GOSTR341112_256_WITH_MAGMA_MGM_L,
		Name: "TLS_GOSTR341112_256_WITH_MAGMA_MGM_L",
		KeySize: 32, IVSize: 8, TagSize: 8,
		Hash: 256, Cipher: CipherMagma, Mode: ModeMGM_L,
	},
	TLS_GOSTR341112_512_WITH_KUZNYECHIK_CTR_OMAC: {
		ID: TLS_GOSTR341112_512_WITH_KUZNYECHIK_CTR_OMAC,
		Name: "TLS_GOSTR341112_512_WITH_KUZNYECHIK_CTR_OMAC",
		KeySize: 32, IVSize: 8, TagSize: 16,
		Hash: 512, Cipher: CipherKuznyechik, Mode: ModeCTR_OMAC,
	},
	TLS_GOSTR341112_512_WITH_KUZNYECHIK_MGM_L: {
		ID: TLS_GOSTR341112_512_WITH_KUZNYECHIK_MGM_L,
		Name: "TLS_GOSTR341112_512_WITH_KUZNYECHIK_MGM_L",
		KeySize: 32, IVSize: 16, TagSize: 16,
		Hash: 512, Cipher: CipherKuznyechik, Mode: ModeMGM_L,
	},
}

// SupportedCipherSuites returns the list of supported GOST cipher suite IDs.
func SupportedCipherSuites() []uint16 {
	return []uint16{
		TLS_GOSTR341112_256_WITH_KUZNYECHIK_MGM_L,
		TLS_GOSTR341112_256_WITH_KUZNYECHIK_CTR_OMAC,
		TLS_GOSTR341112_256_WITH_MAGMA_MGM_L,
		TLS_GOSTR341112_512_WITH_KUZNYECHIK_MGM_L,
	}
}

// SupportedSignatureAlgorithms returns supported GOST signature algorithms.
func SupportedSignatureAlgorithms() []uint16 {
	return []uint16{
		SignatureGOSTR34102012_256,
		SignatureGOSTR34102012_512,
	}
}

// SupportedCurves returns supported GOST curves.
func SupportedCurves() []uint16 {
	return []uint16{
		CurveGOSTR34102012_256_A,
		CurveGOSTR34102012_512_A,
		CurveGOSTR34102012_512_B,
		CurveGOSTR34102012_512_C,
	}
}
