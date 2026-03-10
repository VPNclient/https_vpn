package all

import (
	"github.com/xtls/xray-core/main/commands/base"
)

var cmdGOST = &base.Command{
	UsageLine: "{{.Exec}} gost [-i <private_key>] [-c <curve_type>]",
	Short:     "Generate GOST 2012-256/512 key pair",
	Long: `
Generate GOST 2012-256 or GOST 2012-512 key pair.

Arguments:

	-i
		The base64 encoded private key (optional).
	-c
		The curve type: 256 or 512 (default: 256).
`,
}

func init() {
	cmdGOST.Run = executeGOST // break init loop
}

var gostInputStr = cmdGOST.Flag.String("i", "", "")
var gostCurveType = cmdGOST.Flag.String("c", "256", "")

func executeGOST(cmd *base.Command, args []string) {
	cmdGost()
} 