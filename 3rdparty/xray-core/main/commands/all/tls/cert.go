package tls

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/protocol/tls/cert"
	"github.com/xtls/xray-core/common/task"
	"github.com/xtls/xray-core/main/commands/base"
	sm2x509 "github.com/tjfoc/gmsm/x509"
	x509 "github.com/xtls/xray-core/common/crypto/x509"
)

// cmdCert is the tls cert command
var cmdCert = &base.Command{
	UsageLine: "{{.Exec}} tls cert [--ca] [--domain=example.com] [--expire=240h] [--algorithm=ecdsa]",
	Short:     "Generate TLS certificates",
	Long: `
Generate TLS certificates.

Arguments:

	-domain=domain_name 
		The domain name for the certificate.

	-name=common_name 
		The common name for the certificate.

	-org=organization 
		The organization name for the certificate.

	-ca 
		Whether this certificate is a CA

	-json 
		The output of certificate to JSON

	-file 
		The certificate path to save.

	-expire 
		Expire time of the certificate. Default value 3 months.

	-algorithm
		The algorithm to use for key generation. Options: ecdsa (default), gost2012_256, gost2012_512, sm2
`,
}

func init() {
	cmdCert.Run = executeCert // break init loop
}

var (
	certDomainNames stringList
	_               = func() bool {
		cmdCert.Flag.Var(&certDomainNames, "domain", "Domain name for the certificate")
		return true
	}()

	certCommonName   = cmdCert.Flag.String("name", "Xray Inc", "The common name of this certificate")
	certOrganization = cmdCert.Flag.String("org", "Xray Inc", "Organization of the certificate")
	certIsCA         = cmdCert.Flag.Bool("ca", false, "Whether this certificate is a CA")
	certJSONOutput   = cmdCert.Flag.Bool("json", true, "Print certificate in JSON format")
	certFileOutput   = cmdCert.Flag.String("file", "", "Save certificate in file.")
	certExpire       = cmdCert.Flag.Duration("expire", time.Hour*24*90 /* 90 days */, "Time until the certificate expires. Default value 3 months.")
	certAlgorithm    = cmdCert.Flag.String("algorithm", "ecdsa", "The algorithm to use for key generation")
)

func executeCert(cmd *base.Command, args []string) {
	var generatedCert *cert.Certificate
	var err error

	// Generate certificate based on algorithm
	switch strings.ToLower(*certAlgorithm) {
	case "gost2012_256", "gost2012_512", "sm2":
		// Use SM2 options for GOST and SM2 algorithms
		var sm2Opts []cert.SM2Option
		if *certIsCA {
			sm2Opts = append(sm2Opts, cert.SM2Authority(*certIsCA))
			sm2Opts = append(sm2Opts, cert.SM2KeyUsage(sm2x509.KeyUsageCertSign|sm2x509.KeyUsageKeyEncipherment|sm2x509.KeyUsageDigitalSignature))
		}

		sm2Opts = append(sm2Opts, cert.SM2NotAfter(time.Now().Add(*certExpire)))
		sm2Opts = append(sm2Opts, cert.SM2CommonName(*certCommonName))
		if len(certDomainNames) > 0 {
			sm2Opts = append(sm2Opts, cert.SM2DNSNames(certDomainNames...))
		}
		sm2Opts = append(sm2Opts, cert.SM2Organization(*certOrganization))

		switch strings.ToLower(*certAlgorithm) {
		case "gost2012_256":
			generatedCert, err = cert.GenerateGOST2012_256(nil, sm2Opts...)
		case "gost2012_512":
			generatedCert, err = cert.GenerateGOST2012_512(nil, sm2Opts...)
		case "sm2":
			generatedCert, err = cert.GenerateSM2(nil, sm2Opts...)
		}
	case "ecdsa", "default":
		// Use standard options for ECDSA algorithm
		var opts []cert.Option
		if *certIsCA {
			opts = append(opts, cert.Authority(*certIsCA))
			opts = append(opts, cert.KeyUsage(x509.KeyUsageCertSign|x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature))
		}

		opts = append(opts, cert.NotAfter(time.Now().Add(*certExpire)))
		opts = append(opts, cert.CommonName(*certCommonName))
		if len(certDomainNames) > 0 {
			opts = append(opts, cert.DNSNames(certDomainNames...))
		}
		opts = append(opts, cert.Organization(*certOrganization))

		generatedCert, err = cert.Generate(nil, opts...)
	default:
		base.Fatalf("unsupported algorithm: %s. Supported algorithms: ecdsa, gost2012_256, gost2012_512, sm2", *certAlgorithm)
	}

	if err != nil {
		base.Fatalf("failed to generate TLS certificate: %s", err)
	}

	if *certJSONOutput {
		printJSON(generatedCert)
	}

	if len(*certFileOutput) > 0 {
		if err := printFile(generatedCert, *certFileOutput); err != nil {
			base.Fatalf("failed to save file: %s", err)
		}
	}
}

func printJSON(certificate *cert.Certificate) {
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

func writeFile(content []byte, name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return common.Error2(f.Write(content))
}

func printFile(certificate *cert.Certificate, name string) error {
	certPEM, keyPEM := certificate.ToPEM()
	return task.Run(context.Background(), func() error {
		return writeFile(certPEM, name+".crt")
	}, func() error {
		return writeFile(keyPEM, name+".key")
	})
}

type stringList []string

func (l *stringList) String() string {
	return "String list"
}

func (l *stringList) Set(v string) error {
	if v == "" {
		base.Fatalf("empty value")
	}
	*l = append(*l, v)
	return nil
}

type jsonCert struct {
	Certificate []string `json:"certificate"`
	Key         []string `json:"key"`
}
