// VKO (Key Agreement) per GOST R 34.10-2012.
// Used for TLS key exchange with GOST.
package tls

import (
	"crypto/cipher"
	"errors"
	"math/big"

	"github.com/nativemind/https-vpn/crypto/ru/gost"
)

// VKO computes shared key using GOST R 34.10-2012 key agreement.
// ukm is the User Keying Material (random value from handshake)
// priv is our private key
// pubX, pubY is the peer's public key point
// Returns the shared key (x-coordinate of the computed point, hashed)
func VKO(curve *gost.Curve, priv *big.Int, pubX, pubY *big.Int, ukm []byte) ([]byte, error) {
	if !curve.IsOnCurve(pubX, pubY) {
		return nil, errors.New("vko: public key not on curve")
	}

	// Convert UKM to big.Int
	ukmInt := new(big.Int).SetBytes(ukm)
	if ukmInt.Sign() == 0 {
		ukmInt = big.NewInt(1)
	}

	// Compute scalar k = (ukm * priv) mod n
	k := new(big.Int).Mul(ukmInt, priv)
	k.Mod(k, curve.N)

	// Compute point K = k * Q (peer's public key)
	kx, ky := curve.ScalarMult(pubX, pubY, k.Bytes())

	// Check for point at infinity
	if kx.Sign() == 0 && ky.Sign() == 0 {
		return nil, errors.New("vko: computed point at infinity")
	}

	// Return x-coordinate as the shared secret
	// In practice, this is often hashed, but for TLS PRF, raw bytes are used
	byteLen := (curve.BitSize + 7) / 8
	result := make([]byte, byteLen)
	kxBytes := kx.Bytes()
	copy(result[byteLen-len(kxBytes):], kxBytes)

	return result, nil
}

// VKO256 performs VKO with 256-bit curve.
func VKO256(priv *big.Int, pubX, pubY *big.Int, ukm []byte) ([]byte, error) {
	return VKO(gost.TC26256A(), priv, pubX, pubY, ukm)
}

// VKO512 performs VKO with 512-bit curve (paramSetA).
func VKO512(priv *big.Int, pubX, pubY *big.Int, ukm []byte) ([]byte, error) {
	return VKO(gost.TC26512A(), priv, pubX, pubY, ukm)
}

// TLSKDF derives TLS keys from the shared secret using GOST PRF.
// masterSecret is the VKO output
// label is the TLS PRF label (e.g., "key expansion")
// seed is the concatenation of client_random and server_random
// length is the desired output length
func TLSKDF(masterSecret, label, seed []byte, length int, hashSize int) []byte {
	// GOST TLS uses a PRF based on HMAC-Streebog
	// PRF(secret, label, seed) = P_<hash>(secret, label + seed)
	// P_hash(secret, seed) = HMAC_hash(secret, A(1) + seed) +
	//                        HMAC_hash(secret, A(2) + seed) + ...
	// A(0) = seed
	// A(i) = HMAC_hash(secret, A(i-1))

	labelSeed := append(label, seed...)
	var result []byte
	var a []byte

	// A(0) = labelSeed
	a = labelSeed

	for len(result) < length {
		// A(i) = HMAC(secret, A(i-1))
		if hashSize == 512 {
			a = gost.HMAC512(masterSecret, a)
		} else {
			a = gost.HMAC256(masterSecret, a)
		}

		// P(i) = HMAC(secret, A(i) + labelSeed)
		data := append(a, labelSeed...)
		var p []byte
		if hashSize == 512 {
			p = gost.HMAC512(masterSecret, data)
		} else {
			p = gost.HMAC256(masterSecret, data)
		}
		result = append(result, p...)
	}

	return result[:length]
}

// DeriveTrafficKeys derives encryption keys from master secret.
// Returns clientWriteKey, serverWriteKey, clientWriteIV, serverWriteIV
func DeriveTrafficKeys(masterSecret, clientRandom, serverRandom []byte, suite CipherSuiteInfo) (
	clientKey, serverKey, clientIV, serverIV []byte) {

	seed := append(serverRandom, clientRandom...)
	keyMaterial := TLSKDF(masterSecret, []byte("key expansion"), seed,
		2*suite.KeySize+2*suite.IVSize, suite.Hash)

	offset := 0
	clientKey = keyMaterial[offset : offset+suite.KeySize]
	offset += suite.KeySize
	serverKey = keyMaterial[offset : offset+suite.KeySize]
	offset += suite.KeySize
	clientIV = keyMaterial[offset : offset+suite.IVSize]
	offset += suite.IVSize
	serverIV = keyMaterial[offset : offset+suite.IVSize]

	return
}

// ComputeMasterSecret computes the TLS master secret from pre-master secret.
func ComputeMasterSecret(preMasterSecret, clientRandom, serverRandom []byte, hashSize int) []byte {
	seed := append(clientRandom, serverRandom...)
	// Master secret is 48 bytes for TLS 1.2
	return TLSKDF(preMasterSecret, []byte("master secret"), seed, 48, hashSize)
}

// ComputeVerifyData computes the Finished message verify_data.
func ComputeVerifyData(masterSecret []byte, label string, handshakeHash []byte, hashSize int) []byte {
	// verify_data = PRF(master_secret, finished_label, Hash(handshake_messages))[0..11]
	return TLSKDF(masterSecret, []byte(label), handshakeHash, 12, hashSize)
}

// CreateRecordCipher creates an AEAD cipher for record encryption.
func CreateRecordCipher(key []byte, suite CipherSuiteInfo) (cipher.AEAD, error) {
	var block cipher.Block
	var err error

	if suite.Cipher == CipherKuznyechik {
		block, err = gost.NewKuznyechik(key)
	} else {
		block, err = gost.NewMagma(key)
	}
	if err != nil {
		return nil, err
	}

	return gost.NewMGM(block)
}
