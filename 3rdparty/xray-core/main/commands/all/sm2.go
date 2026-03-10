package all

import (
	"encoding/base64"
	"fmt"
	"os"

	sm2crypto "github.com/xtls/xray-core/common/crypto/sm2"
	"github.com/xtls/xray-core/main/commands/base"
	"github.com/tjfoc/gmsm/sm2"
)

var cmdSM2 = &base.Command{
	UsageLine: "{{.Exec}} sm2 [-i <private_key>]",
	Short:     "Generate SM2 key pair",
	Long: `
Generate SM2 key pair.

Arguments:

	-i
		The base64 encoded private key (optional).
`,
}

func init() {
	cmdSM2.Run = executeSM2 // break init loop
}

var sm2InputStr = cmdSM2.Flag.String("i", "", "")

func executeSM2(cmd *base.Command, args []string) {
	cmdSM2Func()
}

func cmdSM2Func() {
	var (
		inputFile = *sm2InputStr
	)

	var privKey *sm2.PrivateKey
	var err error

	if inputFile != "" {
		// Read private key from file
		data, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read private key file: %v\n", err)
			os.Exit(1)
		}
		privKey, err = sm2crypto.ParsePrivateKey(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse private key: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Generate new key pair
		privKey, err = sm2crypto.GenerateKeyPair()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate SM2 key pair: %v\n", err)
			os.Exit(1)
		}
	}

	// Get public key information
	curveName, pubX, pubY := sm2crypto.GetPublicKeyInfo(privKey)
	privKeyBase64 := base64.StdEncoding.EncodeToString(privKey.D.Bytes())

	fmt.Printf("%s\n", curveName)
	fmt.Printf("Private key: %s\n", privKeyBase64)
	fmt.Printf("Public key X: %s\n", pubX)
	fmt.Printf("Public key Y: %s\n", pubY)
} 