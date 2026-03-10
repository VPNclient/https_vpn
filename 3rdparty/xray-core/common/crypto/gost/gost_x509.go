package gost

import (
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	x509 "github.com/xtls/xray-core/common/crypto/x509"
	pkix "crypto/x509/pkix"
	"github.com/pedroalbanese/gogost/gost3410"
)

// GenerateGOSTSelfSignedCert создает самоподписанный X.509 сертификат GOST2012 (256 или 512)
func GenerateGOSTSelfSignedCert(curve *gost3410.Curve, cn string, expireDays int) ([]byte, []byte, error) {
	fmt.Printf("DEBUG: Starting GenerateGOSTSelfSignedCert\n")
	
	prvRaw := make([]byte, curve.PointSize())
	_, err := rand.Read(prvRaw)
	if err != nil {
		return nil, nil, err
	}
	prv, err := gost3410.NewPrivateKey(curve, prvRaw)
	if err != nil {
		return nil, nil, err
	}
	pub, err := prv.PublicKey()
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("DEBUG: Generated private key type: %T\n", prv)
	fmt.Printf("DEBUG: Generated public key type: %T\n", pub)

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(expireDays) * 24 * time.Hour)
	serial := big.NewInt(time.Now().UnixNano())
	
	// Determine the correct signature algorithm based on curve size
	var signatureAlgorithm x509.SignatureAlgorithm
	if curve.PointSize() == 32 {
		signatureAlgorithm = x509.GOST256
	} else {
		signatureAlgorithm = x509.GOST512
	}
	
	template := x509.Certificate{
		SerialNumber:       serial,
		NotBefore:          notBefore,
		NotAfter:           notAfter,
		SignatureAlgorithm: signatureAlgorithm,
		Subject:            pkix.Name{CommonName: cn},
		KeyUsage:           x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		IsCA: false,
	}

	fmt.Printf("DEBUG: About to call CreateCertificate with signature algorithm: %d\n", signatureAlgorithm)
	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template, &template, pub,
		&gost3410.PrivateKeyReverseDigest{Prv: prv},
	)
	if err != nil {
		fmt.Printf("DEBUG: CreateCertificate failed: %v\n", err)
		return nil, nil, err
	}

	fmt.Printf("DEBUG: CreateCertificate succeeded\n")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalPKCS8PrivateKey(prv)
	if err != nil {
		return nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM, nil
}

// GenerateGOSTCAChildCert создает сертификат, подписанный CA на GOST2012
func GenerateGOSTCAChildCert(curve *gost3410.Curve, sigAlg x509.SignatureAlgorithm, cn string, expireDays int, caCert *x509.Certificate, caKey *gost3410.PrivateKey) ([]byte, []byte, error) {
	prvRaw := make([]byte, curve.PointSize())
	_, err := rand.Read(prvRaw)
	if err != nil {
		return nil, nil, err
	}
	prv, err := gost3410.NewPrivateKey(curve, prvRaw)
	if err != nil {
		return nil, nil, err
	}
	pub, err := prv.PublicKey()
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(expireDays) * 24 * time.Hour)
	serial := big.NewInt(time.Now().UnixNano())
	
	// Determine the correct signature algorithm based on curve size
	var signatureAlgorithm x509.SignatureAlgorithm
	if curve.PointSize() == 32 {
		signatureAlgorithm = x509.GOST256
	} else {
		signatureAlgorithm = x509.GOST512
	}
	
	template := x509.Certificate{
		SerialNumber:       serial,
		NotBefore:          notBefore,
		NotAfter:           notAfter,
		SignatureAlgorithm: signatureAlgorithm,
		Subject:            pkix.Name{CommonName: cn},
		KeyUsage:           x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		IsCA: false,
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template, caCert, pub,
		&gost3410.PrivateKeyReverseDigest{Prv: caKey},
	)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalPKCS8PrivateKey(prv)
	if err != nil {
		return nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM, nil
} 