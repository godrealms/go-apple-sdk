# Changelog

## Unreleased

### Security

- **JWS chain validation**: 所有 JWS 验签路径现在执行完整 RFC 5280 证书链校验，验到内嵌的 Apple Root CA G3，且强制 leaf cert 携带 Apple receipt-signing OID。此前 SDK 仅校验 JWS 签名本身，接受任意 CA 签发的 leaf cert —— 这意味着每个使用 `SignedPayload.DecodedPayload`、`JWSTransaction.Decrypt`、`JWSRenewalInfo.Decrypt` 的接入方都可以被冒充。**影响范围：** 所有消费 App Store Server Notifications V2 或 App Store Server API 返回的 SignedTransactionInfo / SignedRenewalInfo 的代码。

  **必要操作：**

  - 如果你之前依赖 `Decrypt()` / `DecodedPayload()` 接受非 Apple 签名的 payload（比如自签测试桩），改用新的 `DecryptWith(v)` / `DecodedPayloadWith(v)` 方法 + 自定义 `*jws.Verifier`。
  - **`types.X5c.GetPublicKey()` 已删除**。它返回 leaf cert 但不验链，是漏洞的关键点之一。调用它的代码会直接编译失败 —— 迁移到 `jws.Verifier`。

### Added

- 新增顶层包 `github.com/godrealms/go-apple-sdk/jws`：
  - `*Verifier` + `NewVerifier(opts ...Option)`（`WithRootCAs` / `WithRequiredOIDs` / `WithClock`）
  - `VerifyAndDecode[T any](v *Verifier, raw string) (*T, error)` 泛型入口
  - `DefaultVerifier()` 进程级单例（`sync.Once`），内嵌 Apple Root CA G3
  - `*VerificationError` + `ReasonCode` 枚举（`structure` / `chain` / `oid` / `expired` / `signature`）
  - `OIDAppleReceiptSigning` (`1.2.840.113635.100.6.11.1`) 常量；`OIDAppleNotificationSigning` (`1.2.840.113635.100.6.29`) 常量预留但**默认不启用**（待真实 sandbox 通知抓包确认）
- `internal/testchain/` 测试助手：在内存里生成完整 ECDSA 链 + 用 leaf key 签 JWS。仓库内任何包都可以 import，外部不能。
- `scripts/update-root-ca.sh`：从 Apple 官方源刷新 `jws/apple_root_ca_g3.pem`，带 SHA-256 校验。

### Changed

- `JWSTransaction.Decrypt`、`JWSRenewalInfo.Decrypt`、`SignedPayload.DecodedPayload` 失败时返回 `*jws.VerificationError`（仍满足 `error` 接口；用 `errors.As` 解包获取 `Reason`）。只检查 `err != nil` 的旧代码继续工作。
- `types/JWSDecodedHeader.go` 折叠为类型别名：`X5c = jws.X5c`、`JWSDecodedHeader = jws.Header`。仅向前兼容用。

### Removed

- `types/x5c.go`（孤儿副本，`X5c` 类型与 `JWSDecodedHeader.go` 内的小写 `x5c` 类型重复，且仓库内部从未引用）。
- 三处复制粘贴的 `parseSignedPayload`，以及它们附带的 RSA / ASN.1 兜底验签分支。Apple 文档明确这些 payload 用 ES256，兜底逻辑只会扩大攻击面。

### Test infrastructure

- 91.0% jws/ 包测试覆盖率（含 13 项验证矩阵 + 并发 stress test + chain / signature / OID / option 单元测试）。
- 三处 caller 迁移测试（types/JWSTransaction、types/JWSRenewalInfo、app-store-server-notifications/V2）确认 `DefaultVerifier` 拒绝测试链 + 自定义 `Verifier` 接受测试链。
