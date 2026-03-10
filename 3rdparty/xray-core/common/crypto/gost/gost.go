package gost

import (
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
	"crypto/x509/pkix"

	"github.com/pedroalbanese/gogost/gost3410"
	"github.com/pedroalbanese/gogost/gost34112012256"
	"github.com/pedroalbanese/gogost/gost34112012512"
	"github.com/tjfoc/gmsm/sm2"
	sm2x509 "github.com/tjfoc/gmsm/x509"
)

// GOSTCurve represents GOST curve types
type GOSTCurve int

const (
	GOST2012_256 GOSTCurve = iota
	GOST2012_512
)

// String returns the string representation of the GOST curve
func (c GOSTCurve) String() string {
	switch c {
	case GOST2012_256:
		return "GOST2012_256"
	case GOST2012_512:
		return "GOST2012_512"
	default:
		return "Unknown"
	}
}

// GOSTPrivateKey represents a GOST private key
type GOSTPrivateKey struct {
	PrivateKey *gost3410.PrivateKey
	Curve      GOSTCurve
}

// GOSTPublicKey represents a GOST public key
type GOSTPublicKey struct {
	PublicKey *gost3410.PublicKey
	Curve     GOSTCurve
}

// getCurve returns the appropriate GOST curve for the given curve type
func getCurve(curve GOSTCurve) *gost3410.Curve {
	switch curve {
	case GOST2012_256:
		return gost3410.CurveIdtc26gost34102012256paramSetA()
	case GOST2012_512:
		return gost3410.CurveIdtc26gost34102012512paramSetA()
	default:
		return nil
	}
}

// GenerateKeyPair generates a true GOST2012 key pair
func GenerateKeyPair(curve GOSTCurve) (*GOSTPrivateKey, error) {
	gostCurve := getCurve(curve)
	if gostCurve == nil {
		return nil, fmt.Errorf("unsupported GOST curve: %s", curve.String())
	}

	privKey, err := gost3410.GenPrivateKey(gostCurve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GOST %s key: %w", curve.String(), err)
	}

	return &GOSTPrivateKey{
		PrivateKey: privKey,
		Curve:      curve,
	}, nil
}

// GenerateCertificate generates a GOST certificate
func GenerateCertificate(privKey *GOSTPrivateKey, domain string, isCA bool, expireHours int) ([]byte, []byte, error) {
	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(expireHours) * time.Hour)

	// For now, we'll use SM2 certificate template since GOST certificates need special handling
	template := &sm2x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Xray GOST Certificate"},
			CommonName:   domain,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              sm2x509.KeyUsageKeyEncipherment | sm2x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []sm2x509.ExtKeyUsage{sm2x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= sm2x509.KeyUsageCertSign
	} else {
		template.DNSNames = []string{domain}
	}

	// Get GOST public key
	gostPubKey, err := privKey.PrivateKey.PublicKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get GOST public key: %w", err)
	}

	// Convert GOST public key to SM2 format for certificate creation
	// This is a temporary workaround until we implement proper GOST certificate handling
	sm2PubKey := &sm2.PublicKey{
		X: gostPubKey.X,
		Y: gostPubKey.Y,
	}

	// Create a temporary SM2 private key for certificate signing
	tempSM2Key, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temporary SM2 key for certificate: %w", err)
	}

	// Create certificate using SM2 (temporary workaround)
	certDER, err := sm2x509.CreateCertificate(template, template, sm2PubKey, tempSM2Key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GOST certificate: %w", err)
	}

	// Encode certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode GOST private key
	privKeyBytes := privKey.PrivateKey.Raw()
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "GOST PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return certPEM, privKeyPEM, nil
}

// ParsePrivateKey parses a GOST private key from PEM format
func ParsePrivateKey(pemData []byte) (*GOSTPrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try to parse as GOST private key for different curves
	var privKey *gost3410.PrivateKey
	var curve GOSTCurve

	// Try GOST2012_256 first
	gostCurve256 := getCurve(GOST2012_256)
	if gostCurve256 != nil {
		if key, err := gost3410.NewPrivateKey(gostCurve256, block.Bytes); err == nil {
			privKey = key
			curve = GOST2012_256
		}
	}

	// If 256-bit failed, try GOST2012_512
	if privKey == nil {
		gostCurve512 := getCurve(GOST2012_512)
		if gostCurve512 != nil {
			if key, err := gost3410.NewPrivateKey(gostCurve512, block.Bytes); err == nil {
				privKey = key
				curve = GOST2012_512
			}
		}
	}

	if privKey == nil {
		return nil, fmt.Errorf("failed to parse GOST private key")
	}

	return &GOSTPrivateKey{
		PrivateKey: privKey,
		Curve:      curve,
	}, nil
}

// GetPublicKeyInfo returns public key information
func GetPublicKeyInfo(privKey *GOSTPrivateKey) (string, string, string) {
	pubKey, err := privKey.PrivateKey.PublicKey()
	if err != nil {
		return "GOST " + privKey.Curve.String() + "-bit (error getting public key)", "", ""
	}
	return "GOST " + privKey.Curve.String() + "-bit",
		fmt.Sprintf("%x", pubKey.X.Bytes()),
		fmt.Sprintf("%x", pubKey.Y.Bytes())
}

// Sign signs data using GOST2012
func (k *GOSTPrivateKey) Sign(data []byte) ([]byte, error) {
	var hash []byte

	switch k.Curve {
	case GOST2012_256:
		h := gost34112012256.New()
		h.Write(data)
		hash = h.Sum(nil)
	case GOST2012_512:
		h := gost34112012512.New()
		h.Write(data)
		hash = h.Sum(nil)
	default:
		return nil, fmt.Errorf("unsupported GOST curve for signing: %s", k.Curve.String())
	}

	signature, err := k.PrivateKey.Sign(rand.Reader, hash, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with GOST: %w", err)
	}

	return signature, nil
}

// Verify verifies a GOST signature
func (k *GOSTPublicKey) Verify(data, signature []byte) bool {
	var hash []byte

	switch k.Curve {
	case GOST2012_256:
		h := gost34112012256.New()
		h.Write(data)
		hash = h.Sum(nil)
	case GOST2012_512:
		h := gost34112012512.New()
		h.Write(data)
		hash = h.Sum(nil)
	default:
		return false
	}

	valid, err := k.PublicKey.VerifyDigest(hash, signature)
	return err == nil && valid
}

// ToSM2Key converts GOST private key to SM2 format (for compatibility)
func (k *GOSTPrivateKey) ToSM2Key() *sm2.PrivateKey {
	pubKey, err := k.PrivateKey.PublicKey()
	if err != nil {
		return nil
	}

	// Convert raw bytes to big.Int
	rawBytes := k.PrivateKey.Raw()
	d := new(big.Int).SetBytes(rawBytes)

	return &sm2.PrivateKey{
		D: d,
		PublicKey: sm2.PublicKey{
			X: pubKey.X,
			Y: pubKey.Y,
		},
	}
} 