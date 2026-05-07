# 中国商用密码（SM 系列）用于 HTTPS VPN

本目录实现了面向中国国家/行业密码体系的加密能力（通常称为 **SM 系列**），用于本项目的 “国密 VPN” 场景：在 HTTPS/TLS（以及证书体系）中使用 SM2/SM3/SM4 等算法进行握手、认证与数据保护。

## 已实现的算法与代码位置

- **SM2（椭圆曲线公钥密码）**：`crypto/cn/sm2/`
  - 代码：`sm2.go`, `curve.go`
- **SM3（哈希）**：`crypto/cn/sm3/`
  - 代码：`sm3.go`
- **SM4（分组密码）**：`crypto/cn/sm4/`
  - 代码：`sm4.go` 以及模式实现 `gcm.go`, `ccm.go`
- **SM9（基于身份的密码，当前用于研究/实验）**：`crypto/cn/sm9/`
  - 代码：`sm9.go` 及配对/域运算相关实现

## TLS 密码套件（Cipher Suites）

本项目在 `crypto/cn/tls/cipher_suites.go` 中定义了国密 TLS 1.3 密码套件与相关常量（依据 **RFC 8998**）：

| ID | 名称 | AEAD | 哈希 |
|----|------|------|------|
| `0x00C6` | `TLS_SM4_GCM_SM3` | SM4-GCM | SM3 |
| `0x00C7` | `TLS_SM4_CCM_SM3` | SM4-CCM | SM3 |

同时包含：

- **签名算法**：`SignatureSM2_SM3` (`0x0708`)
- **命名曲线**：`CurveSM2` (`0x0029`)

## 如何启用（配置）

在本项目中，`tlsSettings.cipherSuites` 字段被复用为 **“选择国家密码 Provider”** 的入口（仍兼容传统“套件列表字符串”的语义）。要启用国密 Provider，配置为：

```json
{
  "tlsSettings": {
    "cipherSuites": "cn"
  }
}
```

说明：Provider 选择逻辑的规格文档见 `flows/sdd-vpn-https-config-ciphersuites/02-specifications.md`。

## 程序化使用（Provider）

Provider 注册通过 `init()` 自动完成（见 `crypto/cn/provider.go`）。最小示例：

```go
import (
	"github.com/nativemind/https-vpn/crypto"
	_ "github.com/nativemind/https-vpn/crypto/cn" // 注册 "cn" provider
)

provider, ok := crypto.Get("cn")
_ = provider
_ = ok
```

## 兼容性与注意事项

- **Go 标准库限制**：`crypto/cn/provider.go` 中注明，Go 标准库 `crypto/tls` **不原生支持** RFC 8998 的 SM 密码套件；因此对标准 `tls.Config` 的配置更多是“占位/声明能力”，实际启用国密 TLS 往往需要自定义 TLS 实现或集成支持国密的 TLS 栈。
- **证书要求**：若要进行真正的国密握手与认证，通常需要 **SM2 证书**、握手摘要使用 **SM3**、记录层使用 **SM4-GCM/CCM**。

## 参考

- [RFC 8998](https://www.rfc-editor.org/rfc/rfc8998) - ShangMi (SM) Cipher Suites for TLS 1.3
- GM/T 0003（SM2）、GM/T 0004（SM3）、GM/T 0002（SM4）相关标准（按需对照）

