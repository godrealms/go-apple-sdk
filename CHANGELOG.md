# Changelog

本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。自 `v2.0.0` 起模块路径为
`github.com/godrealms/go-apple-sdk/v2`，import 需相应更新。

## v2.0.0 - 2026-07-08

> **破坏性发布**：模块路径迁移到 `github.com/godrealms/go-apple-sdk/v2`。升级请将所有
> import 从 `github.com/godrealms/go-apple-sdk/...` 改为
> `github.com/godrealms/go-apple-sdk/v2/...`，并执行
> `go get github.com/godrealms/go-apple-sdk/v2@v2.0.0`。本版本一次性合并自 `v1.0.0`
> 以来累积的全部破坏性变更（JWS 完整证书链校验、`context.Context` 迁移、
> `GetNotificationHistory` 签名变更等，详见下方）。

### 代码审查修复批次（2026-07）

#### Fixed

- `types.Timestamp.Time()` 此前整除 1000 丢弃毫秒；改用 `time.UnixMilli`，保留 Apple 毫秒精度。
- `types.StorefrontCountryCode.IsValid()` 逻辑反转（`return s == ""`）已修正为 `len(s) == 3`。
- 遗留 `Client.Request` 的 `[]string` 查询参数此前用 `SetQueryParam`（覆盖）只保留最后一个值，
  改用 `QueryParam.Add` 保留全部多值。
- `Client.handleError` 对 3xx 响应此前静默返回 `nil`（调用方误判成功）；现对任何非 2xx 返回错误。
- **`AppStoreServer.GetNotificationHistory`（BREAKING 签名变更）**：此前返回空结构体 `struct{}`
  丢弃全部响应，且把 `paginationToken` 错放进 POST body。现返回填充完整字段的
  `NotificationHistoryResponse`，`paginationToken` 作为查询参数，筛选条件通过新增的
  `NotificationHistoryRequest` 传入 —— 签名变为
  `GetNotificationHistory(ctx, client, NotificationHistoryRequest, paginationToken)`。
- `AppStoreConnect` 的 URL 构造此前未转义调用方 ID，恶意 ID 可注入查询参数/fragment；
  `resolveURL` 现对相对路径段做百分号编码。
- `types.ParsePrivateKey` PKCS#8 分支的 `err` 变量遮蔽已消除（安全关键路径）。
- `NewConfig` 现只在裸 base64 输入上补 PEM 头尾，不再对已带 PEM 头的输入二次包裹。

#### Security

- **凭证脱敏**：`Client.handleError` 的诊断日志此前会打印含 `Authorization: Bearer <JWT>`
  的完整请求头；现对 `Authorization` / `Cookie` / `Proxy-Authorization` 脱敏。
- **并发安全**：根 `Client` 的 `SetService` 此前非同步改写共享状态（数据竞争，`-race` 可检出）。
  改为构造期一次性装配每服务独立的 resty client + 互斥保护的服务选择器；连接池不再被丢弃。
- **供应链**：升级 `golang.org/x/net`→v0.56.0、`golang-jwt/jwt/v5`→v5.3.1，`go.mod` 固定
  `go 1.25` + `toolchain go1.26.4`。`govulncheck ./...` 从 11 个可达漏洞降为 **0**。

#### Added

- `WithServiceBaseURL(service, url)` 客户端 Option：覆盖某服务的 base URL（便于测试/代理/mock）。
- 工程化：GitHub Actions CI（多 OS + `-race` + 覆盖率）、`govulncheck` workflow、Dependabot、
  `SECURITY.md`、`Makefile`；`.golangci.yml` 迁移到 v2 schema（此前因版本不兼容加载失败）。
- 回归测试：为上述每个 bug 补充失败→转绿的回归测试；根包覆盖率 0%→60%、`types` 8.5%→45%、
  `app-store-server` 0%→17%。

### Security

- **JWS chain validation**: 所有 JWS 验签路径现在执行完整 RFC 5280 证书链校验，验到内嵌的 Apple Root CA G3，且强制 leaf cert 携带 Apple receipt-signing OID。此前 SDK 仅校验 JWS 签名本身，接受任意 CA 签发的 leaf cert —— 这意味着每个使用 `SignedPayload.DecodedPayload`、`JWSTransaction.Decrypt`、`JWSRenewalInfo.Decrypt` 的接入方都可以被冒充。**影响范围：** 所有消费 App Store Server Notifications V2 或 App Store Server API 返回的 SignedTransactionInfo / SignedRenewalInfo 的代码。

  **必要操作：**

  - 如果你之前依赖 `Decrypt()` / `DecodedPayload()` 接受非 Apple 签名的 payload（比如自签测试桩），改用新的 `DecryptWith(v)` / `DecodedPayloadWith(v)` 方法 + 自定义 `*jws.Verifier`。
  - **`types.X5c.GetPublicKey()` 已删除**。它返回 leaf cert 但不验链，是漏洞的关键点之一。调用它的代码会直接编译失败 —— 迁移到 `jws.Verifier`。

### Added

- 新增顶层包 `github.com/godrealms/go-apple-sdk/v2/jws`：
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
