package x509

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"io"
	"math/big"
	"net"
	"time"

	"github.com/pedroalbanese/gogost/gost3410"
	"github.com/pedroalbanese/gogost/gost34112012256"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
	"crypto/elliptic"
)

// SignatureAlgorithm represents the algorithm used to sign the certificate
type SignatureAlgorithm int

const (
	UnknownSignatureAlgorithm SignatureAlgorithm = iota
	MD2WithRSA
	MD5WithRSA
	SHA1WithRSA
	SHA256WithRSA
	SHA384WithRSA
	SHA512WithRSA
	DSAWithSHA1
	DSAWithSHA256
	ECDSAWithSHA1
	ECDSAWithSHA256
	ECDSAWithSHA384
	ECDSAWithSHA512
	GOST256
	GOST512
)

// KeyUsage represents the set of operations that are valid for a given key
type KeyUsage int

const (
	KeyUsageDigitalSignature KeyUsage = 1 << iota
	KeyUsageContentCommitment
	KeyUsageKeyEncipherment
	KeyUsageDataEncipherment
	KeyUsageKeyAgreement
	KeyUsageCertSign
	KeyUsageCRLSign
	KeyUsageEncipherOnly
	KeyUsageDecipherOnly
)

// ExtKeyUsage represents an extended set of operations that are valid for a given key
type ExtKeyUsage int

const (
	ExtKeyUsageAny ExtKeyUsage = iota
	ExtKeyUsageServerAuth
	ExtKeyUsageClientAuth
	ExtKeyUsageCodeSigning
	ExtKeyUsageEmailProtection
	ExtKeyUsageIPSECEndSystem
	ExtKeyUsageIPSECTunnel
	ExtKeyUsageIPSECUser
	ExtKeyUsageTimeStamping
	ExtKeyUsageOCSPSigning
	ExtKeyUsageMicrosoftServerGatedCrypto
	ExtKeyUsageNetscapeServerGatedCrypto
	ExtKeyUsageMicrosoftCommercialCodeSigning
	ExtKeyUsageMicrosoftKernelCodeSigning
)

// Certificate represents an X.509 certificate
type Certificate struct {
	Raw                     []byte // Complete ASN.1 DER content (certificate, signature algorithm and signature).
	RawTBSCertificate       []byte // Certificate part of raw ASN.1 DER content.
	RawSubjectPublicKeyInfo []byte // DER encoded SubjectPublicKeyInfo.
	RawSubject              []byte // DER encoded Subject
	RawIssuer               []byte // DER encoded Issuer

	Signature          []byte
	SignatureAlgorithm SignatureAlgorithm

	PublicKeyAlgorithm PublicKeyAlgorithm
	PublicKey          interface{}

	Version             int
	SerialNumber       *big.Int
	Issuer             pkix.Name
	Subject            pkix.Name
	NotBefore, NotAfter time.Time // Validity bounds.
	KeyUsage           KeyUsage

	// Extensions contains raw X.509 extensions. When parsing certificates,
	// this can be used to extract non-critical extensions that are not
	// parsed by this package. When marshaling certificates, the Extensions
	// field is ignored, see ExtraExtensions.
	Extensions []pkix.Extension

	// ExtraExtensions contains extensions to be copied, raw, into any
	// marshaled certificates. Values override any extensions that would
	// otherwise be produced based on the other fields. The ExtraExtensions
	// field is not populated when parsing certificates, see Extensions.
	ExtraExtensions []pkix.Extension

	// UnhandledCriticalExtensions contains a list of extension IDs that
	// were not (fully) processed when parsing. Verify will fail if this
	// slice is non-empty unless using InsecureSkipVerify.
	UnhandledCriticalExtensions []asn1.ObjectIdentifier

	ExtKeyUsage        []ExtKeyUsage
	UnknownExtKeyUsage []asn1.ObjectIdentifier

	BasicConstraintsValid bool
	IsCA                  bool
	MaxPathLen            int
	// MaxPathLenZero indicates that BasicConstraintsValid is true and that
	// MaxPathLen is zero. MaxPathLenZero should be true if and only if
	// MaxPathLen is zero.
	MaxPathLenZero bool

	SubjectKeyId   []byte
	AuthorityKeyId []byte

	// RFC 5280, 4.2.2.1 (Authority Information Access)
	OCSPServer            []string
	IssuingCertificateURL []string

	// Subject Alternate Name values
	DNSNames       []string
	EmailAddresses []string
	IPAddresses    []net.IP

	// Name constraints
	PermittedDNSDomainsCritical bool // if true then the name constraints are marked critical.
	PermittedDNSDomains         []string
	ExcludedDNSDomains          []string
	PermittedIPRanges           []*net.IPNet
	ExcludedIPRanges            []*net.IPNet
	PermittedEmailAddresses     []string
	ExcludedEmailAddresses      []string
	PermittedURIDomains         []string
	ExcludedURIDomains          []string

	// CRL Distribution Points
	CRLDistributionPoints []string

	PolicyIdentifiers []asn1.ObjectIdentifier
}

// PublicKeyAlgorithm represents the algorithm used to sign the certificate
type PublicKeyAlgorithm int

const (
	UnknownPublicKeyAlgorithm PublicKeyAlgorithm = iota
	RSA
	DSA
	ECDSA
	GOST
)

// CreateCertificate creates a new certificate based on a template. The
// template is also used as the parent certificate. The certificate is signed
// using the private key of the parent.
func CreateCertificate(rand io.Reader, template, parent *Certificate, pub interface{}, priv interface{}) ([]byte, error) {
	// Validate inputs
	if template == nil {
		return nil, fmt.Errorf("template cannot be nil")
	}
	if parent == nil {
		parent = template // Self-signed certificate
	}
	if pub == nil {
		return nil, fmt.Errorf("public key cannot be nil")
	}
	if priv == nil {
		return nil, fmt.Errorf("private key cannot be nil")
	}

	// Debug: print private key type
	fmt.Printf("DEBUG: Private key type: %T\n", priv)

	// Determine signature algorithm
	var sigAlg SignatureAlgorithm
	switch priv.(type) {
	case *rsa.PrivateKey:
		sigAlg = SHA256WithRSA
		fmt.Printf("DEBUG: Using RSA signature algorithm\n")
	case *ecdsa.PrivateKey:
		sigAlg = ECDSAWithSHA256
		fmt.Printf("DEBUG: Using ECDSA signature algorithm\n")
	case *sm2.PrivateKey:
		sigAlg = ECDSAWithSHA256
		fmt.Printf("DEBUG: Using SM2 signature algorithm\n")
	case *gost3410.PrivateKey:
		// Determine GOST algorithm based on curve
		gostPriv := priv.(*gost3410.PrivateKey)
		// Get curve size from the private key's raw bytes
		rawBytes := gostPriv.Raw()
		if len(rawBytes) == 32 {
			sigAlg = GOST256
		} else {
			sigAlg = GOST512
		}
		fmt.Printf("DEBUG: Using GOST signature algorithm: %d\n", sigAlg)
	case *gost3410.PrivateKeyReverseDigest:
		// Determine GOST algorithm based on curve
		gostPriv := priv.(*gost3410.PrivateKeyReverseDigest)
		rawBytes := gostPriv.Prv.Raw()
		if len(rawBytes) == 32 {
			sigAlg = GOST256
		} else {
			sigAlg = GOST512
		}
		fmt.Printf("DEBUG: Using GOST ReverseDigest signature algorithm: %d\n", sigAlg)
	default:
		return nil, fmt.Errorf("unsupported private key type: %T", priv)
	}

	// Create TBS certificate
	tbsCert, err := createTBSCertificate(template, pub, sigAlg)
	if err != nil {
		return nil, fmt.Errorf("failed to create TBS certificate: %w", err)
	}

	// Sign the TBS certificate
	signature, err := signCertificate(tbsCert, priv, sigAlg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign certificate: %w", err)
	}

	// Create the final certificate
	cert := &Certificate{
		Raw:                     append(tbsCert, signature...),
		RawTBSCertificate:       tbsCert,
		Signature:               signature,
		SignatureAlgorithm:      sigAlg,
		PublicKeyAlgorithm:      getPublicKeyAlgorithm(pub),
		PublicKey:               pub,
		Version:                 template.Version,
		SerialNumber:            template.SerialNumber,
		Issuer:                  template.Issuer,
		Subject:                 template.Subject,
		NotBefore:               template.NotBefore,
		NotAfter:                template.NotAfter,
		KeyUsage:                template.KeyUsage,
		Extensions:              template.Extensions,
		ExtraExtensions:         template.ExtraExtensions,
		ExtKeyUsage:             template.ExtKeyUsage,
		BasicConstraintsValid:   template.BasicConstraintsValid,
		IsCA:                    template.IsCA,
		MaxPathLen:              template.MaxPathLen,
		MaxPathLenZero:          template.MaxPathLenZero,
		SubjectKeyId:            template.SubjectKeyId,
		AuthorityKeyId:          template.AuthorityKeyId,
		OCSPServer:              template.OCSPServer,
		IssuingCertificateURL:   template.IssuingCertificateURL,
		DNSNames:                template.DNSNames,
		EmailAddresses:          template.EmailAddresses,
		IPAddresses:             template.IPAddresses,
		CRLDistributionPoints:   template.CRLDistributionPoints,
		PolicyIdentifiers:       template.PolicyIdentifiers,
	}

	return cert.Raw, nil
}

// MarshalPKCS8PrivateKey converts a private key to PKCS #8, ASN.1 DER form
func MarshalPKCS8PrivateKey(key interface{}) ([]byte, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return marshalPKCS8PrivateKey(k)
	case *ecdsa.PrivateKey:
		return marshalECDSAPKCS8PrivateKey(k)
	case *sm2.PrivateKey:
		return marshalSM2PKCS8PrivateKey(k)
	case *gost3410.PrivateKey:
		return marshalGOSTPKCS8PrivateKey(k)
	default:
		return nil, fmt.Errorf("unsupported private key type")
	}
}

// ParseCertificate parses a single certificate from the given ASN.1 DER data
func ParseCertificate(der []byte) (*Certificate, error) {
	// This is a simplified implementation
	// In a full implementation, you would parse the ASN.1 DER data
	// and populate the Certificate struct
	return nil, fmt.Errorf("ParseCertificate not implemented")
}

// ParseECPrivateKey parses an EC private key in SEC 1, ASN.1 DER form
func ParseECPrivateKey(der []byte) (*ecdsa.PrivateKey, error) {
	// Simplified implementation
	return nil, fmt.Errorf("ParseECPrivateKey not implemented")
}

// ParsePKCS8PrivateKey parses an unencrypted private key in PKCS #8, ASN.1 DER form
func ParsePKCS8PrivateKey(der []byte) (interface{}, error) {
	// Simplified implementation
	return nil, fmt.Errorf("ParsePKCS8PrivateKey not implemented")
}

// ParsePKCS1PrivateKey parses an RSA private key in PKCS #1, ASN.1 DER form
func ParsePKCS1PrivateKey(der []byte) (*rsa.PrivateKey, error) {
	// Simplified implementation
	return nil, fmt.Errorf("ParsePKCS1PrivateKey not implemented")
}

// Helper functions
func createTBSCertificate(template *Certificate, pub interface{}, sigAlg SignatureAlgorithm) ([]byte, error) {
	// Create a basic ASN.1 DER encoding of the To-Be-Signed certificate
	// This is a simplified implementation for testing purposes
	
	fmt.Printf("DEBUG: createTBSCertificate called with signature algorithm: %d\n", sigAlg)
	
	// Determine OIDs based on signature algorithm
	var signatureOID asn1.ObjectIdentifier
	var publicKeyOID asn1.ObjectIdentifier
	
	switch sigAlg {
	case SHA256WithRSA:
		signatureOID = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11} // SHA256WithRSA
		publicKeyOID = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1} // RSA
	case ECDSAWithSHA256:
		signatureOID = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2} // ECDSAWithSHA256
		publicKeyOID = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1} // EC
	case GOST256:
		signatureOID = asn1.ObjectIdentifier{1, 2, 643, 7, 1, 1, 3, 2} // GOST R 34.10-2012 256-bit
		publicKeyOID = asn1.ObjectIdentifier{1, 2, 643, 7, 1, 1, 1, 1} // GOST R 34.10-2012 256-bit
	case GOST512:
		signatureOID = asn1.ObjectIdentifier{1, 2, 643, 7, 1, 1, 3, 3} // GOST R 34.10-2012 512-bit
		publicKeyOID = asn1.ObjectIdentifier{1, 2, 643, 7, 1, 1, 1, 2} // GOST R 34.10-2012 512-bit
	default:
		signatureOID = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2} // Default to ECDSAWithSHA256
		publicKeyOID = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1} // EC
	}
	
	// Encode the public key based on its type
	var publicKeyBytes []byte
	var bitLength int
	
	switch p := pub.(type) {
	case *rsa.PublicKey:
		// For RSA, we need to encode the modulus and exponent
		publicKeyBytes = []byte{0x04} // Placeholder for RSA public key
		bitLength = p.N.BitLen()
	case *ecdsa.PublicKey:
		// For ECDSA, encode the point in uncompressed format
		publicKeyBytes = append([]byte{0x04}, p.X.Bytes()...)
		publicKeyBytes = append(publicKeyBytes, p.Y.Bytes()...)
		bitLength = len(publicKeyBytes) * 8
	case *sm2.PublicKey:
		// For SM2, encode the point in uncompressed format
		publicKeyBytes = append([]byte{0x04}, p.X.Bytes()...)
		publicKeyBytes = append(publicKeyBytes, p.Y.Bytes()...)
		bitLength = len(publicKeyBytes) * 8
	case *gost3410.PublicKey:
		// For GOST, encode the point in uncompressed format (like ECDSA)
		publicKeyBytes = append([]byte{0x04}, p.X.Bytes()...)
		publicKeyBytes = append(publicKeyBytes, p.Y.Bytes()...)
		bitLength = len(publicKeyBytes) * 8
	default:
		return nil, fmt.Errorf("unsupported public key type: %T", pub)
	}
	
	// For GOST certificates, we need to use a different approach
	// GOST certificates have a specific ASN.1 structure that's different from standard X.509
	if sigAlg == GOST256 || sigAlg == GOST512 {
		fmt.Printf("DEBUG: Using GOST-specific certificate structure\n")
		// Create GOST-specific certificate structure
		gostCert := struct {
			Version            int `asn1:"optional,explicit,default:0,tag:0"`
			SerialNumber       *big.Int
			SignatureAlgorithm pkix.AlgorithmIdentifier
			Issuer            pkix.Name
			Validity          struct {
				NotBefore time.Time
				NotAfter  time.Time
			}
			Subject            pkix.Name
			SubjectPublicKeyInfo struct {
				Algorithm pkix.AlgorithmIdentifier
				PublicKey asn1.BitString
			}
			IssuerUniqueID  asn1.BitString `asn1:"optional,tag:1"`
			SubjectUniqueID asn1.BitString `asn1:"optional,tag:2"`
			Extensions      []pkix.Extension `asn1:"optional,tag:3"`
		}{
			Version:      template.Version,
			SerialNumber: template.SerialNumber,
			SignatureAlgorithm: pkix.AlgorithmIdentifier{
				Algorithm: signatureOID,
			},
			Issuer: template.Issuer,
			Validity: struct {
				NotBefore time.Time
				NotAfter  time.Time
			}{
				NotBefore: template.NotBefore,
				NotAfter:  template.NotAfter,
			},
			Subject: template.Subject,
			SubjectPublicKeyInfo: struct {
				Algorithm pkix.AlgorithmIdentifier
				PublicKey asn1.BitString
			}{
				Algorithm: pkix.AlgorithmIdentifier{
					Algorithm: publicKeyOID,
				},
				PublicKey: asn1.BitString{
					Bytes:     publicKeyBytes,
					BitLength: bitLength,
				},
			},
			IssuerUniqueID:  asn1.BitString{},
			SubjectUniqueID: asn1.BitString{},
			Extensions:      template.Extensions,
		}

		// Encode to ASN.1 DER
		der, err := asn1.Marshal(gostCert)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal GOST TBS certificate: %w", err)
		}

		fmt.Printf("DEBUG: GOST TBS certificate marshaled successfully, length: %d\n", len(der))
		return der, nil
	}
	
	fmt.Printf("DEBUG: Using standard X.509 certificate structure\n")
	// Create the basic certificate structure with proper ASN.1 tags
	tbs := struct {
		Version            int `asn1:"optional,explicit,default:0,tag:0"`
		SerialNumber       *big.Int
		SignatureAlgorithm pkix.AlgorithmIdentifier
		Issuer            pkix.Name
		Validity          struct {
			NotBefore time.Time
			NotAfter  time.Time
		}
		Subject            pkix.Name
		SubjectPublicKeyInfo struct {
			Algorithm pkix.AlgorithmIdentifier
			PublicKey asn1.BitString
		}
		IssuerUniqueID  asn1.BitString `asn1:"optional,tag:1"`
		SubjectUniqueID asn1.BitString `asn1:"optional,tag:2"`
		Extensions      []pkix.Extension `asn1:"optional,tag:3"`
	}{
		Version:      template.Version,
		SerialNumber: template.SerialNumber,
		SignatureAlgorithm: pkix.AlgorithmIdentifier{
			Algorithm: signatureOID,
		},
		Issuer: template.Issuer,
		Validity: struct {
			NotBefore time.Time
			NotAfter  time.Time
		}{
			NotBefore: template.NotBefore,
			NotAfter:  template.NotAfter,
		},
		Subject: template.Subject,
		SubjectPublicKeyInfo: struct {
			Algorithm pkix.AlgorithmIdentifier
			PublicKey asn1.BitString
		}{
			Algorithm: pkix.AlgorithmIdentifier{
				Algorithm: publicKeyOID,
			},
			PublicKey: asn1.BitString{
				Bytes:     publicKeyBytes,
				BitLength: bitLength,
			},
		},
		IssuerUniqueID:  asn1.BitString{},
		SubjectUniqueID: asn1.BitString{},
		Extensions:      template.Extensions,
	}

	// Encode to ASN.1 DER
	der, err := asn1.Marshal(tbs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TBS certificate: %w", err)
	}

	return der, nil
}

func signCertificate(tbsCert []byte, priv interface{}, sigAlg SignatureAlgorithm) ([]byte, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return signWithRSA(tbsCert, k, sigAlg)
	case *ecdsa.PrivateKey:
		return signWithECDSA(tbsCert, k, sigAlg)
	case *sm2.PrivateKey:
		return signWithSM2(tbsCert, k, sigAlg)
	case *gost3410.PrivateKey:
		return signWithGOST(tbsCert, k, sigAlg)
	case *gost3410.PrivateKeyReverseDigest:
		return signWithGOSTReverseDigest(tbsCert, k, sigAlg)
	default:
		return nil, fmt.Errorf("unsupported private key type for signing")
	}
}

func signWithRSA(data []byte, priv *rsa.PrivateKey, sigAlg SignatureAlgorithm) ([]byte, error) {
	// Simplified RSA signing
	return nil, fmt.Errorf("RSA signing not implemented")
}

func signWithECDSA(data []byte, priv *ecdsa.PrivateKey, sigAlg SignatureAlgorithm) ([]byte, error) {
	// Hash the data with SHA256 for ECDSA signing
	hash := sha256.Sum256(data)
	
	// Sign the hash
	signature, err := ecdsa.SignASN1(rand.Reader, priv, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign with ECDSA: %w", err)
	}
	
	return signature, nil
}

func signWithSM2(data []byte, priv *sm2.PrivateKey, sigAlg SignatureAlgorithm) ([]byte, error) {
	// Hash the data with SM3 for SM2 signing
	hash := sm3.Sm3Sum(data)
	
	// Sign the hash
	signature, err := priv.Sign(rand.Reader, hash, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with SM2: %w", err)
	}
	
	return signature, nil
}

func signWithGOST(data []byte, priv *gost3410.PrivateKey, sigAlg SignatureAlgorithm) ([]byte, error) {
	// Use GOST signing with reverse digest
	signer := &gost3410.PrivateKeyReverseDigest{Prv: priv}
	return signer.Sign(rand.Reader, data, nil)
}

func signWithGOSTReverseDigest(data []byte, priv *gost3410.PrivateKeyReverseDigest, sigAlg SignatureAlgorithm) ([]byte, error) {
	// Use GOST signing with reverse digest
	// First, we need to hash the data with GOST hash function
	hash := gost34112012256.New()
	hash.Write(data)
	hashed := hash.Sum(nil)
	
	// Sign the hash
	signature, err := priv.Sign(rand.Reader, hashed, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with GOST: %w", err)
	}
	
	return signature, nil
}

func getPublicKeyAlgorithm(pub interface{}) PublicKeyAlgorithm {
	switch pub.(type) {
	case *rsa.PublicKey:
		return RSA
	case *ecdsa.PublicKey:
		return ECDSA
	case *sm2.PublicKey:
		return ECDSA
	case *gost3410.PublicKey:
		return GOST
	default:
		return UnknownPublicKeyAlgorithm
	}
}

func marshalPKCS8PrivateKey(key *rsa.PrivateKey) ([]byte, error) {
	// Simplified PKCS8 marshaling for RSA
	return nil, fmt.Errorf("RSA PKCS8 marshaling not implemented")
}

func marshalECDSAPKCS8PrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	// Create PKCS8 structure for ECDSA private key
	pkcs8 := struct {
		Version    int
		Algorithm  pkix.AlgorithmIdentifier
		PrivateKey []byte
	}{
		Version: 0,
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm: asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}, // EC algorithm
		},
		PrivateKey: key.D.Bytes(),
	}

	// Encode to ASN.1 DER
	der, err := asn1.Marshal(pkcs8)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ECDSA PKCS8 private key: %w", err)
	}

	return der, nil
}

func marshalSM2PKCS8PrivateKey(key *sm2.PrivateKey) ([]byte, error) {
	// Create PKCS8 structure for SM2 private key
	pkcs8 := struct {
		Version    int
		Algorithm  pkix.AlgorithmIdentifier
		PrivateKey []byte
	}{
		Version: 0,
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm: asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}, // EC algorithm (SM2 uses EC format)
		},
		PrivateKey: key.D.Bytes(),
	}

	// Encode to ASN.1 DER
	der, err := asn1.Marshal(pkcs8)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SM2 PKCS8 private key: %w", err)
	}

	return der, nil
}

func marshalGOSTPKCS8PrivateKey(key *gost3410.PrivateKey) ([]byte, error) {
	// Determine the correct OID based on the curve size
	var algorithmOID asn1.ObjectIdentifier
	rawKey := key.Raw()
	
	if len(rawKey) == 32 {
		// GOST R 34.10-2012 256-bit
		algorithmOID = asn1.ObjectIdentifier{1, 2, 643, 7, 1, 1, 1, 1}
	} else {
		// GOST R 34.10-2012 512-bit
		algorithmOID = asn1.ObjectIdentifier{1, 2, 643, 7, 1, 1, 1, 2}
	}
	
	// Create PKCS8 structure for GOST private key
	pkcs8 := struct {
		Version    int
		Algorithm  pkix.AlgorithmIdentifier
		PrivateKey []byte
	}{
		Version: 0,
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm: algorithmOID,
		},
		PrivateKey: rawKey,
	}

	// Encode to ASN.1 DER
	der, err := asn1.Marshal(pkcs8)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GOST PKCS8 private key: %w", err)
	}

	return der, nil
} 

// encodeSubjectPublicKeyInfo формирует ASN.1 DER SubjectPublicKeyInfo и возвращает AlgorithmIdentifier и BitString
func encodeSubjectPublicKeyInfo(pub interface{}, sigAlg SignatureAlgorithm) (pkix.AlgorithmIdentifier, asn1.BitString, error) {
	var algo pkix.AlgorithmIdentifier
	var pubKeyBytes []byte
	var bitLength int

	switch p := pub.(type) {
	case *rsa.PublicKey:
		algo = pkix.AlgorithmIdentifier{
			Algorithm: asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1},
		}
		type rsaPub struct {
			N *big.Int
			E int
		}
		pubASN1, err := asn1.Marshal(rsaPub{p.N, p.E})
		if err != nil {
			return algo, asn1.BitString{}, err
		}
		pubKeyBytes = pubASN1
		bitLength = len(pubKeyBytes) * 8
	case *ecdsa.PublicKey:
		curveOID := asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7} // secp256r1
		curveOIDBytes, _ := asn1.Marshal(curveOID)
		algo = pkix.AlgorithmIdentifier{
			Algorithm: asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1},
			Parameters: asn1.RawValue{
				FullBytes: curveOIDBytes,
			},
		}
		pubKeyBytes = ellipticMarshal(p.Curve, p.X, p.Y)
		bitLength = len(pubKeyBytes) * 8
	case *sm2.PublicKey:
		sm2OID := asn1.ObjectIdentifier{1, 2, 156, 10197, 1, 301, 1}
		sm2OIDBytes, _ := asn1.Marshal(sm2OID)
		algo = pkix.AlgorithmIdentifier{
			Algorithm: sm2OID,
			Parameters: asn1.RawValue{
				FullBytes: sm2OIDBytes,
			},
		}
		pubKeyBytes = ellipticMarshal(p.Curve, p.X, p.Y)
		bitLength = len(pubKeyBytes) * 8
	case *gost3410.PublicKey:
		gostOID := asn1.ObjectIdentifier{1, 2, 643, 7, 1, 1, 1, 1}
		gostOIDBytes, _ := asn1.Marshal(gostOID)
		algo = pkix.AlgorithmIdentifier{
			Algorithm: gostOID,
			Parameters: asn1.RawValue{FullBytes: gostOIDBytes},
		}
		pubKeyBytes = append([]byte{0x04}, p.X.Bytes()...)
		pubKeyBytes = append(pubKeyBytes, p.Y.Bytes()...)
		bitLength = len(pubKeyBytes) * 8
	default:
		return algo, asn1.BitString{}, fmt.Errorf("unsupported public key type: %T", pub)
	}
	return algo, asn1.BitString{Bytes: pubKeyBytes, BitLength: bitLength}, nil
}

// ellipticMarshal — стандартная функция для маршалинга точки на кривой
func ellipticMarshal(curve elliptic.Curve, x, y *big.Int) []byte {
	return append([]byte{0x04}, append(x.Bytes(), y.Bytes()...)...)
} 