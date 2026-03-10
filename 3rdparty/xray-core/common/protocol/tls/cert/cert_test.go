package cert

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/task"
)

func TestGenerate(t *testing.T) {
	err := generate(nil, true, true, "ca")
	if err != nil {
		t.Fatal(err)
	}
}

func generate(domainNames []string, isCA bool, jsonOutput bool, fileOutput string) error {
	commonName := "Xray Root CA"
	organization := "Xray Inc"

	expire := time.Hour * 3

	var opts []Option
	if isCA {
		opts = append(opts, Authority(isCA))
		opts = append(opts, KeyUsage(x509.KeyUsageCertSign|x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature))
	}

	opts = append(opts, NotAfter(time.Now().Add(expire)))
	opts = append(opts, CommonName(commonName))
	if len(domainNames) > 0 {
		opts = append(opts, DNSNames(domainNames...))
	}
	opts = append(opts, Organization(organization))

	cert, err := Generate(nil, opts...)
	if err != nil {
		return errors.New("failed to generate TLS certificate").Base(err)
	}

	if jsonOutput {
		printJSON(cert)
	}

	if len(fileOutput) > 0 {
		if err := printFile(cert, fileOutput); err != nil {
			return err
		}
	}

	return nil
}

type jsonCert struct {
	Certificate []string `json:"certificate"`
	Key         []string `json:"key"`
}

func printJSON(certificate *Certificate) {
	certPEM, keyPEM := certificate.ToPEM()
	jCert := &jsonCert{
		Certificate: strings.Split(strings.TrimSpace(string(certPEM)), "\n"),
		Key:         strings.Split(strings.TrimSpace(string(keyPEM)), "\n"),
	}
	content, err := json.MarshalIndent(jCert, "", "  ")
	common.Must(err)
	os.Stdout.Write(content)
	os.Stdout.WriteString("\n")
}

func printFile(certificate *Certificate, name string) error {
	certPEM, keyPEM := certificate.ToPEM()
	return task.Run(context.Background(), func() error {
		return writeFile(certPEM, name+".crt")
	}, func() error {
		return writeFile(keyPEM, name+".key")
	})
}

func writeFile(content []byte, name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return common.Error2(f.Write(content))
}

func TestGenerateGOST2012_256(t *testing.T) {
	cert, err := GenerateGOST2012_256(nil, CommonName("test.gost256.local"), DNSNames("test.gost256.local"))
	if err != nil {
		t.Fatal("Failed to generate GOST 2012-256 certificate:", err)
	}

	if len(cert.Certificate) == 0 {
		t.Fatal("Generated certificate is empty")
	}

	if len(cert.PrivateKey) == 0 {
		t.Fatal("Generated private key is empty")
	}

	// Parse the certificate to verify it's valid
	x509Cert, err := x509.ParseCertificate(cert.Certificate)
	if err != nil {
		t.Fatal("Failed to parse generated certificate:", err)
	}

	if x509Cert.Subject.CommonName != "test.gost256.local" {
		t.Fatal("Certificate CommonName mismatch")
	}

	if !x509Cert.NotAfter.After(time.Now()) {
		t.Fatal("Certificate is already expired")
	}
}

func TestGenerateGOST2012_512(t *testing.T) {
	cert, err := GenerateGOST2012_512(nil, CommonName("test.gost512.local"), DNSNames("test.gost512.local"))
	if err != nil {
		t.Fatal("Failed to generate GOST 2012-512 certificate:", err)
	}

	if len(cert.Certificate) == 0 {
		t.Fatal("Generated certificate is empty")
	}

	if len(cert.PrivateKey) == 0 {
		t.Fatal("Generated private key is empty")
	}

	// Parse the certificate to verify it's valid
	x509Cert, err := x509.ParseCertificate(cert.Certificate)
	if err != nil {
		t.Fatal("Failed to parse generated certificate:", err)
	}

	if x509Cert.Subject.CommonName != "test.gost512.local" {
		t.Fatal("Certificate CommonName mismatch")
	}

	if !x509Cert.NotAfter.After(time.Now()) {
		t.Fatal("Certificate is already expired")
	}
}

func TestGenerateSM2(t *testing.T) {
	cert, err := GenerateSM2(nil, CommonName("test.sm2.local"), DNSNames("test.sm2.local"))
	if err != nil {
		t.Fatal("Failed to generate SM2 certificate:", err)
	}

	if len(cert.Certificate) == 0 {
		t.Fatal("Generated certificate is empty")
	}

	if len(cert.PrivateKey) == 0 {
		t.Fatal("Generated private key is empty")
	}

	// Parse the certificate to verify it's valid
	x509Cert, err := x509.ParseCertificate(cert.Certificate)
	if err != nil {
		t.Fatal("Failed to parse generated certificate:", err)
	}

	if x509Cert.Subject.CommonName != "test.sm2.local" {
		t.Fatal("Certificate CommonName mismatch")
	}

	if !x509Cert.NotAfter.After(time.Now()) {
		t.Fatal("Certificate is already expired")
	}
}

func TestMustGenerateGOST2012_256(t *testing.T) {
	cert := MustGenerateGOST2012_256(nil, CommonName("test.gost256.must.local"), DNSNames("test.gost256.must.local"))
	
	if len(cert.Certificate) == 0 {
		t.Fatal("Generated certificate is empty")
	}

	if len(cert.PrivateKey) == 0 {
		t.Fatal("Generated private key is empty")
	}
}

func TestMustGenerateGOST2012_512(t *testing.T) {
	cert := MustGenerateGOST2012_512(nil, CommonName("test.gost512.must.local"), DNSNames("test.gost512.must.local"))
	
	if len(cert.Certificate) == 0 {
		t.Fatal("Generated certificate is empty")
	}

	if len(cert.PrivateKey) == 0 {
		t.Fatal("Generated private key is empty")
	}
}

func TestMustGenerateSM2(t *testing.T) {
	cert := MustGenerateSM2(nil, CommonName("test.sm2.must.local"), DNSNames("test.sm2.must.local"))
	
	if len(cert.Certificate) == 0 {
		t.Fatal("Generated certificate is empty")
	}

	if len(cert.PrivateKey) == 0 {
		t.Fatal("Generated private key is empty")
	}
}
