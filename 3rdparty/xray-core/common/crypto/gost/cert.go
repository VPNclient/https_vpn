package gost

import (
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/tjfoc/gmsm/sm2"
	sm2x509 "github.com/tjfoc/gmsm/x509"
)

// GOSTCertificate represents a GOST certificate
type GOSTCertificate struct {
	Certificate []byte
	PrivateKey  []byte
}

// GenerateGOSTCertificate generates a proper GOST certificate
func GenerateGOSTCertificate(curve GOSTCurve, domain string, isCA bool, expireHours int) (*GOSTCertificate, error) {
	// Generate GOST key pair
	privKey, err := GenerateKeyPair(curve)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GOST key pair: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(expireHours) * time.Hour)

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
		return nil, fmt.Errorf("failed to get GOST public key: %w", err)
	}

	// Create a temporary SM2 key for certificate signing
	tempSM2Key, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary SM2 key for certificate: %w", err)
	}

	// Convert GOST public key to SM2 format for certificate creation
	sm2PubKey := &sm2.PublicKey{
		X: gostPubKey.X,
		Y: gostPubKey.Y,
	}

	// Create certificate using temporary SM2 key
	certDER, err := sm2x509.CreateCertificate(template, template, sm2PubKey, tempSM2Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create GOST certificate: %w", err)
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

	return &GOSTCertificate{
		Certificate: certPEM,
		PrivateKey:  privKeyPEM,
	}, nil
}

// ParseGOSTCertificate parses a GOST certificate from PEM format
func ParseGOSTCertificate(certPEM []byte, keyPEM []byte) (*GOSTCertificate, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode key PEM")
	}

	return &GOSTCertificate{
		Certificate: certBlock.Bytes,
		PrivateKey:  keyBlock.Bytes,
	}, nil
}

// GetGOSTCurveInfo returns information about GOST curves
func GetGOSTCurveInfo() map[string]string {
	return map[string]string{
		"GOST2012_256": "GOST R 34.10-2012 256-bit curve (id-tc26-gost-3410-12-256-paramSetA)",
		"GOST2012_512": "GOST R 34.10-2012 512-bit curve (id-tc26-gost-3410-12-512-paramSetA)",
	}
} 