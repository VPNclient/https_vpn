// HMAC with Streebog hash per GOST R 34.11-2012.
package gost

import (
	"crypto/hmac"
	"hash"
)

// NewHMAC256 returns a new HMAC-Streebog-256.
func NewHMAC256(key []byte) hash.Hash {
	return hmac.New(NewStreebog256, key)
}

// NewHMAC512 returns a new HMAC-Streebog-512.
func NewHMAC512(key []byte) hash.Hash {
	return hmac.New(NewStreebog512, key)
}

// HMAC256 computes HMAC-Streebog-256 of data with key.
func HMAC256(key, data []byte) []byte {
	h := NewHMAC256(key)
	h.Write(data)
	return h.Sum(nil)
}

// HMAC512 computes HMAC-Streebog-512 of data with key.
func HMAC512(key, data []byte) []byte {
	h := NewHMAC512(key)
	h.Write(data)
	return h.Sum(nil)
}
