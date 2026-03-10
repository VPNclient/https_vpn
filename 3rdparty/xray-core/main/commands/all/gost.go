package all

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/xtls/xray-core/common/crypto/gost"
)

func cmdGost() {
	var (
		curveType = flag.String("c", "256", "GOST curve type (256 or 512)")
		inputFile = flag.String("i", "", "Input private key file")
	)
	flag.Parse()

	var curve gost.GOSTCurve
	switch *curveType {
	case "256":
		curve = gost.GOST2012_256
	case "512":
		curve = gost.GOST2012_512
	default:
		fmt.Fprintf(os.Stderr, "Unsupported curve type: %s. Supported: 256, 512\n", *curveType)
		os.Exit(1)
	}

	var privKey *gost.GOSTPrivateKey
	var err error

	if *inputFile != "" {
		// Read private key from file
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read private key file: %v\n", err)
			os.Exit(1)
		}
		privKey, err = gost.ParsePrivateKey(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse private key: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Generate new key pair
		privKey, err = gost.GenerateKeyPair(curve)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate GOST key pair: %v\n", err)
			os.Exit(1)
		}
	}

	// Get public key information
	curveName, pubX, pubY := gost.GetPublicKeyInfo(privKey)
	privKeyBase64 := base64.StdEncoding.EncodeToString(privKey.PrivateKey.Raw())

	fmt.Printf("%s\n", curveName)
	fmt.Printf("Private key: %s\n", privKeyBase64)
	fmt.Printf("Public key X: %s\n", pubX)
	fmt.Printf("Public key Y: %s\n", pubY)
} 