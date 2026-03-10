package sm2crypto

import (
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
	"crypto/x509/pkix"

	"github.com/tjfoc/gmsm/sm2"
	sm2x509 "github.com/tjfoc/gmsm/x509"
)

// GenerateKeyPair generates an SM2 key pair
func GenerateKeyPair() (*sm2.PrivateKey, error) {
	privKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SM2 key: %w", err)
	}
	return privKey, nil
}

// GenerateCertificate generates an SM2 certificate
func GenerateCertificate(privKey *sm2.PrivateKey, domain string, isCA bool, expireHours int) ([]byte, []byte, error) {
	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(expireHours) * time.Hour)

	template := &sm2x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Xray SM2 Certificate"},
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

	// Create certificate using SM2
	certDER, err := sm2x509.CreateCertificate(template, template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create SM2 certificate: %w", err)
	}

	// Encode certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key in PKCS8 format for better compatibility
	privKeyBytes, err := sm2x509.MarshalSm2UnecryptedPrivateKey(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal SM2 private key: %w", err)
	}

	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return certPEM, privKeyPEM, nil
}

// ParsePrivateKey parses an SM2 private key from PEM format
func ParsePrivateKey(pemData []byte) (*sm2.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privKey, err := sm2x509.ParsePKCS8UnecryptedPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SM2 private key: %w", err)
	}

	return privKey, nil
}

// GetPublicKeyInfo returns public key information
func GetPublicKeyInfo(privKey *sm2.PrivateKey) (string, string, string) {
	pubKey := privKey.PublicKey
	return "SM2",
		fmt.Sprintf("%x", pubKey.X.Bytes()),
		fmt.Sprintf("%x", pubKey.Y.Bytes())
} 