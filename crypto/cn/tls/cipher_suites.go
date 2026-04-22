// Package tls defines TLS cipher suites for Chinese cryptography per RFC 8998.
package tls

// SM TLS 1.3 cipher suites (RFC 8998)
const (
	TLS_SM4_GCM_SM3 uint16 = 0x00C6
	TLS_SM4_CCM_SM3 uint16 = 0x00C7
)

// SM2 signature algorithm
const (
	SignatureSM2_SM3 uint16 = 0x0708
)

// SM2 named curve
const (
	CurveSM2 uint16 = 0x0029
)
