# JWS 通知 / 交易验签：补完整 Chain Validation 设计

- **状态**：Draft（待 review）
- **日期**：2026-05-05
- **范围**：仓库内所有 JWS 验签路径（App Store Server Notifications V2 + JWSTransaction.Decrypt + JWSRenewalInfo.Decrypt）
- **优先级**：CRITICAL（安全 bug 修复）
- **关联**：仓库整体审查（2026-05-04 brainstorm 会话）拆出的 sub-project A。同批次还有 B（Gen 1 → Gen 2 收敛）/ C（Phase 6.2 前置）/ D（types/ 重组）/ E（小修小补打包）—— 不在本文档范围。

---

## 1. 问题陈述

### 1.1 当前缺陷

仓库目前所有 JWS 验签路径都**只验证签名本身、不验证证书链回溯到可信根**。这意味着任何持有合法 CA 签发的 leaf cert 的攻击者，都可以伪造 App Store Server Notifications 或 SignedTransactionInfo / SignedRenewalInfo 让 SDK 接受。

涉及文件：

| 文件 | 行 | 问题 |
|---|---|---|
| `types/x5c.go` | 17–34 | `X5c.GetPublicKey()` 只 parse `x5c[0]` 返回，不验链 |
| `types/JWSDecodedHeader.go` | 15–32 | 同上的小写 `x5c.GetPublicKey()`，又是一份重复实现 |
| `types/JWSTransaction.go` | 147–192 | `Decrypt()` 调上面两个之一 → 链未验证 + 兼容 RSA / ASN.1 fallback 扩大攻击面 |
| `types/JWSRenewalInfo.go` | 106–151 | 同 JWSTransaction，几乎逐字复制 |
| `app-store-server-notifications/App.Store.Server.Notifications.V2.go` | 99–144 | `DecodedPayload()` 同模式：链未验证 + RSA fallback + ASN.1 fallback |

### 1.2 影响

- 任何用本 SDK 处理 App Store Server Notifications 的接入方都受影响。本质上 SDK 没有兑现"通知是 Apple 签发的"这个核心安全承诺
- 攻击场景：攻击者控制能签发任意 cert 的 CA（含被攻陷的 public CA 或 mis-issued cert），即可伪造一份带合法签名的 JWS 让 SDK 接受
- 重复代码（三处 `parseSignedPayload`、两份 `X5c` 类型）放大了维护成本：未来即便修了一处，另两处仍埋雷

### 1.3 与 Apple 官方实现的差距

Apple 在 Java/Swift 官方 SDK 里做的是 **完整 RFC 5280 chain validation + 检查 leaf cert 上的 Apple receipt-signing OID 家族**（`1.2.840.113635.100.6.11.x` 等）。本 SDK 应该至少对齐这条标准。

---

## 2. 目标 / 非目标

### 2.1 目标

- 所有 JWS 验签点统一走完整 RFC 5280 chain validation，验到 Apple Root CA G3
- 强制检查 leaf cert 上的 Apple WWDR OID
- 抽出共享的 `jws/` 包，消除三处 `parseSignedPayload` 和两份 `X5c` 重复
- 现有 `Decrypt()` / `DecodedPayload()` API 签名不变，调用方升级 SDK 即获得安全升级
- 新增显式 `*Verifier` API 供需要自定义 root pool 的调用方（测试 / 未来根证书轮换）

### 2.2 非目标

- ❌ **CRL / OCSP revocation 检查**：引入网络往返与可用性依赖，得不偿失
- ❌ **签发 JWS**：本 SDK 是消费者
- ❌ **SPKI pinning**：被 chain validation 取代
- ❌ **多 root 自动轮替**：调用方可通过 `WithRootCAs` 自己组合
- ❌ **缓存解析后的 leaf cert**：每次解析的开销可忽略

---

## 3. 决策日志（来自 brainstorm Q1–Q5）

| # | 决策 | 选择 |
|---|---|---|
| Q1 | 验签严格度 | 完整 chain validation + 强制 OID 检查（最严，对齐 Apple 官方） |
| Q2 | 修复范围 | 所有 JWS 路径 + 合并两份重复的 X5c 类型 |
| Q3 | 信任锚来源 | 默认 embed Apple Root CA G3，允许 `WithRootCAs` 覆盖 |
| Q4 | 现有 API 兼容性 | 保留现有方法签名 + 内部用默认 verifier；新增显式 `DecryptWith(*Verifier)` |
| Q5 | 代码归属 | 新建顶层包 `jws/` |

---

## 4. 架构

### 4.1 包文件布局

```
jws/
├── doc.go                       // 包级文档
├── verifier.go                  // *Verifier、NewVerifier、Option
├── verifier_test.go
├── chain.go                     // chain validation：x509.Verify 包装
├── oid.go                       // Apple OID 常量 + leaf cert OID 检查
├── x5c.go                       // X5c 类型 + 解析（合并两份重复）
├── header.go                    // Header 结构（迁自 types/JWSDecodedHeader.go）
├── decode.go                    // VerifyAndDecode[T] 泛型入口
├── decode_test.go
├── default.go                   // sync.Once 懒初始化 DefaultVerifier
├── apple_root_ca_g3.pem         // //go:embed 的官方 root（PEM）
├── internal/
│   └── testchain/
│       └── testchain.go         // 测试夹具：即时构造伪 ECDSA 链
└── testdata/
    └── real_apple_notification.txt   // sandbox 抓的真实通知（一次写入，回归用）
```

### 4.2 公共 API 表面

```go
// jws/verifier.go

// Verifier 持有 root CA pool、required OID 列表、clock。安全用于并发。
type Verifier struct { /* unexported fields */ }

// NewVerifier 构造一个 Verifier。无 Option 时使用所有默认值（embedded G3 + Apple OIDs + time.Now）。
func NewVerifier(opts ...Option) *Verifier

// Option 是 Verifier 构造选项。
type Option func(*Verifier)

// WithRootCAs 用调用方提供的 pool 完全替换默认 G3 root。
// 用于测试桩或未来 Apple 轮换 root 时调用方先打补丁。
func WithRootCAs(pool *x509.CertPool) Option

// WithRequiredOIDs 替换默认 OID 列表。leaf cert 至少包含其中一个才能通过验证。
func WithRequiredOIDs(oids ...asn1.ObjectIdentifier) Option

// WithClock 注入时钟，用于测试过期 / 未生效证书。生产代码不应使用。
func WithClock(now func() time.Time) Option

// DefaultVerifier 返回进程级共享的默认 Verifier（embed Apple Root CA G3 + 默认 OID）。
// 第一次调用时 sync.Once 初始化；解析 embedded PEM 失败时 panic（不应发生）。
func DefaultVerifier() *Verifier
```

```go
// jws/x5c.go

// X5c 是 JWS x5c header 数组（每个元素是 base64-DER 编码的证书）。
type X5c []string

// Parse 把 x5c 数组解码为 []*x509.Certificate。chain[0] 是 leaf。
func (x X5c) Parse() ([]*x509.Certificate, error)
```

```go
// jws/header.go

// Header 对应 JWS 的 protected header（仅含 SDK 关心的字段）。
type Header struct {
    Alg string `json:"alg"`
    X5c X5c    `json:"x5c"`
}

// Alg 是 JWS 算法字符串。当前 SDK 只接受 "ES256"。
type Alg = string
```

```go
// jws/decode.go

// VerifyAndDecode 是统一入口：拆 JWS、验签、检 OID、反序列化 payload。
//
// 失败时返回 *VerificationError（不会返回其他类型的 error）。
func VerifyAndDecode[T any](v *Verifier, raw string) (*T, error)
```

```go
// jws/verifier.go (continued)

// VerificationError 是 jws 包返回的唯一错误类型。
type VerificationError struct {
    Reason ReasonCode
    Cause  error  // 可为 nil
}

func (e *VerificationError) Error() string
func (e *VerificationError) Unwrap() error

// ReasonCode 区分失败种类，便于调用方做 alerting / metrics。
type ReasonCode int

const (
    ReasonStructure ReasonCode = iota + 1  // JWS 格式错（含 alg ≠ ES256）
    ReasonChain                            // 证书链验证失败（unknown authority 等）
    ReasonOID                              // leaf cert 缺少 required OID
    ReasonExpired                          // 证书过期或未生效
    ReasonSignature                        // 签名验证失败（含长度错）
)

func (r ReasonCode) String() string  // "structure" / "chain" / "oid" / "expired" / "signature"
```

```go
// jws/oid.go

// Apple WWDR OID 常量。
var (
    // OIDAppleReceiptSigning 是 Apple Receipt Signing OID。
    // Apple StoreKit 文档已确认 receipt-signing leaf cert 必含此 OID。
    OIDAppleReceiptSigning = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 11, 1}

    // OIDAppleNotificationSigning 是 App Store Server Notifications V2 OID。
    // ⚠️ TBD：此 OID 是否真为 V2 通知 leaf 所用，需用真实 sandbox 抓包夹具核对。
    OIDAppleNotificationSigning = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 29}
)

// DefaultRequiredOIDs 是 DefaultVerifier 使用的 OID 列表（任一即可）。
var DefaultRequiredOIDs = []asn1.ObjectIdentifier{
    OIDAppleReceiptSigning,
    OIDAppleNotificationSigning,
}
```

### 4.3 与 `types/` 的关系

`types/JWSDecodedHeader.go` 简化为：

```go
package types

import "github.com/godrealms/go-apple-sdk/jws"

// 向后兼容的类型别名 —— 让外部 import "types" 的代码仍能编译
type Alg = jws.Alg
type X5c = jws.X5c
type JWSDecodedHeader = jws.Header
```

`types/x5c.go` **整个删除**。这是孤儿副本（仓库内部从未被引用），但即便外部代码在 import `types.X5c`，也会经由别名指向 `jws.X5c`。

**有意的破坏性变更**：`X5c.GetPublicKey()` 不再存在。任何还在调那个不安全方法的外部代码会编译失败 —— 这正是修复目的之一，强制迁移到 `jws.Verifier`。

---

## 5. 验证流程

`VerifyAndDecode[T]` 的实现步骤：

1. **拆 JWS**
   - `strings.Split(raw, ".")` → 必须 3 段
   - 每段 base64.RawURLEncoding 解码
   - 任一失败 → `&VerificationError{Reason: ReasonStructure, Cause: ...}`

2. **解码 header；强制 alg == "ES256"**
   - `json.Unmarshal(headerBytes, &Header{})`
   - `header.Alg != "ES256"` → `ReasonStructure`

3. **解析 x5c → []*x509.Certificate**
   - 每个元素 base64.StdEncoding 解码 + `x509.ParseCertificate`
   - 链长度 < 1 或任一解析失败 → `ReasonStructure`

4. **chain validation**
   ```go
   intermediates := x509.NewCertPool()
   for _, cert := range chain[1:] {
       intermediates.AddCert(cert)
   }
   _, err := chain[0].Verify(x509.VerifyOptions{
       Roots:         v.roots,
       Intermediates: intermediates,
       CurrentTime:   v.clock(),
   })
   ```
   - 错误映射：
     - `x509.CertificateInvalidError` 且 `Reason==Expired || Reason==NotYetValid` → `ReasonExpired`
     - 其他所有 `x509.*Error`（含 `UnknownAuthorityError`） → `ReasonChain`

5. **OID 检查**
   - 遍历 `chain[0].Extensions`（`[]pkix.Extension`），把每个 `Extension.Id` 与 `v.requiredOIDs` 比对
   - 至少匹配一个 → 通过；否则 → `ReasonOID`
   - 注：用 `Extensions` 而不是 `UnknownExtKeyUsage`，因为 Apple 的 receipt-signing OID 通常注册为 cert 自定义 extension 而非 EKU

6. **签名验证**（**先验签再 unmarshal payload**，避免在未验证数据上跑 JSON 解析器）
   - 强制 `len(sigBytes) == 64`（ES256 / P-256 IEEE P1363 raw）
   - 长度不对 → `ReasonSignature`
   - `r := big.Int.SetBytes(sigBytes[:32])`，`s := big.Int.SetBytes(sigBytes[32:])`
   - `hash := sha256.Sum256([]byte(headerB64 + "." + payloadB64))`
   - `pub, ok := chain[0].PublicKey.(*ecdsa.PublicKey)`；不是 ECDSA → `ReasonSignature`
   - `ecdsa.Verify(pub, hash[:], r, s)` 失败 → `ReasonSignature`

7. **解码 payload → T**（仅在签名通过之后）
   - `json.Unmarshal(payloadBytes, out)`
   - 失败 → `ReasonStructure`

8. **返回解码后的 \*T**

### 关键决定（重申）

- **强制 ES256**：legacy 的 RSA / ASN.1 兜底全部删除。Apple 文档明确 receipt 与通知用 ES256
- **没有 RSA 路径**：`case *rsa.PublicKey` 删除
- **没有 ASN.1 fallback**：`ecdsa.VerifyASN1` 二次尝试删除。JWS 规范规定 ES256 是 raw 64 字节
- **`x509.Verify` 已包含完整 RFC 5280 路径构建**：包括 SAN / Basic Constraints / Path Length / 签名算法兼容性等检查

### 性能

- 单次 `VerifyAndDecode` 的成本：~3 次 base64 解码 + 1 次 JSON unmarshal（header）+ 1 次 chain validation（含 1–2 次 ECDSA 验证）+ 1 次 leaf 签名验证 + 1 次 JSON unmarshal（payload）
- 基线 ~1ms，无 I/O
- `DefaultVerifier()` 用 `sync.Once`：embedded PEM 仅首次解析；之后所有 goroutine 共享同一个 `*Verifier`
- `*Verifier` 字段全部不可变（`roots *x509.CertPool` 一旦构造不变；`requiredOIDs` 只读 slice）—— 显式标注 "safe for concurrent use"

---

## 6. 错误模型示例

```go
payload, err := jws.VerifyAndDecode[Transaction](v, raw)
if err != nil {
    var verErr *jws.VerificationError
    if errors.As(err, &verErr) {
        switch verErr.Reason {
        case jws.ReasonExpired:
            metrics.IncrCounter("apple.jws.cert_expired")
            // 一般是 SDK 升级延迟 —— 让 ops 知道
        case jws.ReasonChain:
            log.Errorf("possible attack or misconfig: %v", verErr)
        case jws.ReasonStructure:
            log.Warnf("malformed JWS from upstream: %v", verErr)
        }
    }
    return err
}
```

错误信息格式：`jws: <reason>: <cause>` —— 例：`jws: chain: x509: certificate signed by unknown authority`

---

## 7. 迁移路径（按文件）

### 7.1 新增

- `jws/` 整个目录（按 4.1 布局）

### 7.2 修改

**`types/JWSTransaction.go`**：
```go
func (j JWSTransaction) Decrypt() (*JWSTransactionDecodedPayload, error) {
    return jws.VerifyAndDecode[JWSTransactionDecodedPayload](jws.DefaultVerifier(), string(j))
}
func (j JWSTransaction) DecryptWith(v *jws.Verifier) (*JWSTransactionDecodedPayload, error) {
    return jws.VerifyAndDecode[JWSTransactionDecodedPayload](v, string(j))
}
// parseSignedPayload 整段删除 (~30 LOC)
// import 清理：crypto / ecdsa / rsa / sha256 / base64 / strings / big 全部移除
```

**`types/JWSRenewalInfo.go`**：同上模式，添加 `Decrypt` + `DecryptWith` 走 `jws.VerifyAndDecode`，删除 `parseSignedPayload`。

**`types/JWSDecodedHeader.go`**：缩成只有别名（见 4.3）。

**`app-store-server-notifications/App.Store.Server.Notifications.V2.go`**：
```go
func (sp SignedPayload) DecodedPayload() (*ResponseBodyV2DecodedPayload, error) {
    return jws.VerifyAndDecode[ResponseBodyV2DecodedPayload](jws.DefaultVerifier(), string(sp))
}
func (sp SignedPayload) DecodedPayloadWith(v *jws.Verifier) (*ResponseBodyV2DecodedPayload, error) {
    return jws.VerifyAndDecode[ResponseBodyV2DecodedPayload](v, string(sp))
}
// parseSignedPayload 整段删除
```

### 7.3 删除

- `types/x5c.go` —— 整文件，孤儿副本
- 三处复制粘贴的 `parseSignedPayload`（每处 ~30 LOC）

### 7.4 净代码量

- 删除：~120 LOC（3 份 parseSignedPayload + 错误验签逻辑 + types/x5c.go）
- 新增：~250 LOC（jws/ 实现 + tests）
- 净 +130 LOC，其中 ~150 是测试代码

---

## 8. 测试策略

### 8.1 设计原则

不依赖 Apple 真实证书做单元测试 —— 真实 leaf 会过期、CI 重现性差、签名材料不在我们手里。

改用**测试时即时构造完整伪 ECDSA 链**：root → intermediate → leaf，全部用 `crypto/ecdsa` 在内存里生成，每次测试一份新的。

### 8.2 测试 helper

```go
// jws/internal/testchain/testchain.go

type TestChain struct {
    Root, Intermediate, Leaf *x509.Certificate
    LeafKey                  *ecdsa.PrivateKey
    RootPool                 *x509.CertPool
}

type Opt func(*config)

func WithLeafOIDs(oids ...asn1.ObjectIdentifier) Opt
func WithLeafNotAfter(t time.Time) Opt
func WithLeafNotBefore(t time.Time) Opt

// New 构造一条 3 级 ECDSA 链（root P-384 / intermediate P-256 / leaf P-256）。
func New(t *testing.T, opts ...Opt) *TestChain

// SignJWS 用 leaf key 签一份合法 JWS（ES256），payload 是任意可 JSON marshal 的对象。
func (tc *TestChain) SignJWS(t *testing.T, payload any) string
```

### 8.3 测试矩阵

| # | 用例 | 期望 Reason |
|---|---|---|
| 1 | 完整合法链 + 合法签名 + payload | nil（成功，且解码后等于原 payload） |
| 2 | 用未注册的 root 签发的链 | `ReasonChain` |
| 3 | leaf cert `NotAfter` < verifier.clock() | `ReasonExpired` |
| 4 | leaf cert `NotBefore` > verifier.clock() | `ReasonExpired` |
| 5 | leaf cert 无 Apple OID extension | `ReasonOID` |
| 6 | payload 被改一字节 | `ReasonSignature` |
| 7 | signature 被截断到 60 字节 | `ReasonSignature` |
| 8 | header.alg == "RS256" | `ReasonStructure` |
| 9 | JWS 只有 2 段 | `ReasonStructure` |
| 10 | x5c 数组为空 | `ReasonStructure` |
| 11 | x5c 包含坏 base64 | `ReasonStructure` |
| 12 | leaf 的 publicKey 是 RSA（不是 ECDSA） | `ReasonSignature` |
| 13 | 真实 sandbox notification + production G3 | nil（用 `testdata/real_apple_notification.txt`，作为锁定回归） |
| 14 | 并发：100 goroutine 同时调 `DefaultVerifier()` 并验签 | 全部成功（验证 `sync.Once` + 共享 verifier 安全） |

### 8.4 覆盖率目标

`jws/` 包测试覆盖率 ≥ 90%，与 `app-store-connect/` 阶段五的 80% 基线对齐并加严（因为是安全敏感代码）。

---

## 9. 风险与发布说明

### 9.1 主要风险

1. **行为破坏（不是 API 破坏）**：调用方 import 路径与方法签名都不变，但 `Decrypt()` 现在会**真的拒**坏证书。如果生产里有调用方误信了恶意 / 过期 / 伪造 JWS 而旧 SDK"凑巧"接受了，升级后这些请求开始返回 `ReasonChain` / `ReasonExpired` 错误。**这正是修复目的**，但 release notes 必须显眼写明。

2. **`X5c.GetPublicKey()` 故意不再存在**：直接编译失败。是有意为之，目的是让所有"自己拿 leaf cert 自己 verify 但跳过 chain"的旧代码停摆。release notes 给迁移示例（用 `jws.Verifier` 替代）。

3. **OID 检查可能误伤**：`OIDAppleNotificationSigning = 6.29` 是否真为 V2 通知 leaf 所用尚未用真实数据核实（见 §11 TBD）。如果不是，Server Notifications 验签会全失败。**实施前必须先核**。

### 9.2 Release Notes 模板

```
BREAKING (behavior, not API):

  go-apple-sdk now performs full RFC 5280 certificate chain validation
  on App Store Server Notifications, signed transactions, and signed
  renewal info. Previously the SDK only verified the JWS signature
  itself and did not check that the certificate was issued by Apple.

  Action required:
  - If you were relying on Decrypt() / DecodedPayload() to succeed on
    payloads signed by anything other than Apple's real leaf certs
    (e.g., test fixtures with self-signed certs), pass a custom
    *jws.Verifier via the new DecryptWith(v) / DecodedPayloadWith(v)
    methods.
  - The deprecated types.X5c.GetPublicKey() method has been removed.
    Migrate to jws.Verifier (see migration guide).
```

### 9.3 embedded G3 PEM 更新流程

新增 `scripts/update-root-ca.sh`：

```bash
#!/usr/bin/env bash
set -euo pipefail
URL="https://www.apple.com/appleca/AppleIncRootCertificate.pem"
EXPECTED_SHA256="<known good hash>"
TMP=$(mktemp)
curl -fsSL "$URL" -o "$TMP"
ACTUAL=$(shasum -a 256 "$TMP" | awk '{print $1}')
if [[ "$ACTUAL" != "$EXPECTED_SHA256" ]]; then
  echo "SHA256 mismatch: expected $EXPECTED_SHA256, got $ACTUAL" >&2
  exit 1
fi
mv "$TMP" jws/apple_root_ca_g3.pem
echo "Updated."
```

Apple Root CA G3 在 2039 到期，本 SDK 应早于此 EOL。

---

## 10. 落地分阶段（高层）

详细的实施步骤在后续 plan 文档里。这里只列高层阶段，便于评估工作量：

1. **阶段 1**：建 `jws/` 包骨架 + `Verifier` + `VerifyAndDecode[T]` + `internal/testchain` + 完整测试矩阵（约 1 天）
2. **阶段 2**：迁移三个 caller（types/JWSTransaction、types/JWSRenewalInfo、app-store-server-notifications/V2），删除重复代码，types/JWSDecodedHeader.go 改为别名（约 0.5 天）
3. **阶段 3**：embedded PEM + `scripts/update-root-ca.sh` + 真实通知夹具锁定（约 0.5 天）
4. **阶段 4**：release notes + CHANGELOG + 例子代码（约 0.5 天）

**预计总工作量**：~2.5 天。`go build ./...` 与 `go test -race ./...` 全绿且 `jws/` coverage ≥ 90% 是验收线。

---

## 11. 开放问题（TBD）

### 🚧 阻塞实施

- **OID `1.2.840.113635.100.6.29` 是否真为 App Store Server Notifications V2 leaf cert 所用**？需要用真实 sandbox 抓的通知 + `openssl x509 -text -noout -in <leaf>.pem` 核对 leaf 的 X.509 extensions 部分列出的 OID。如果不是，§4.2 的 `OIDAppleNotificationSigning` 与 `DefaultRequiredOIDs` 必须调整 —— 否则 SDK 会把所有合法 V2 通知拒掉。**写代码前必须解决**。

### 📋 实施过程中决定即可

- **真实 sandbox notification 夹具从哪来**？跑 sandbox 抓一次 + redact 敏感字段后落到 `jws/testdata/real_apple_notification.txt`。可以在阶段 3 实施时收尾。
- **examples/ 是否要改**？现有 examples 的 `Decrypt()` 调用自动获益（API 签名不变）。但建议在阶段 4 给 1–2 个示例加 `errors.As(*VerificationError)` 的错误处理片段，演示新的错误分类。

---

## 12. 不会做的事（明文）

- **不**增加 CRL / OCSP revocation 检查
- **不**新增签发 JWS 的能力
- **不**做 SPKI pinning
- **不**自动下载 / 轮换 root CA
- **不**支持非 Apple 的 JWS 验签（虽然 jws 包 *可以* 通过 `WithRootCAs` 复用，但不在文档承诺范围内）
- **不**在本 sub-project 范围内修复 [体检报告](../notes/2026-05-04-repo-audit-findings.md)（如已落地）的其他 HIGH 级问题（例如 client.go:66 死分支、PUT 调 GET 端点 bug）—— 那些归 sub-project B/C/D/E
