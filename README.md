# Go Apple SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/godrealms/go-apple-sdk.svg)](https://pkg.go.dev/github.com/godrealms/go-apple-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/godrealms/go-apple-sdk)](https://goreportcard.com/report/github.com/godrealms/go-apple-sdk)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Go Apple SDK 是一个覆盖 Apple 开发者生态两套核心 API 的 Go SDK，帮助后端服务和 CI/CD 流水线自动化应用内购买校验、订阅生命周期管理、TestFlight 发布、App Store 提审与商品配置等任务。

本仓库的功能分成两大模块：

- **App Store Server API**：服务器端处理订阅和 IAP 相关的事务，包括交易验证、订阅状态、退款记录、消费信息、以及 App Store Server Notifications v2 的解析。
- **App Store Connect API**：开发者自动化，覆盖 Apps / Reports / TestFlight / Provisioning / Users / App Store 提审与元数据 / In-App Purchases v2 / Subscription Groups 等主要资源。完整实现路线图见 [APP_STORE_CONNECT_API_PLAN.md](./APP_STORE_CONNECT_API_PLAN.md)。

## 功能特性

- 类型安全的泛型 JSON:API 解码（`Document[T]` / `Resource[T]` / `Relationship`）
- 强类型流式 Query Builder：`NewQuery().Filter(...).Include(...).Limit(...)`
- 自动游标分页（`Paginator[T]` + `ListAll` / `ListIterator`）
- 统一的错误模型：Apple 返回的业务错误统一映射为 `*APIError`，本地构造/传输错误为 `*ClientError`
- 可选的结构化日志 Hook（`Logger` 接口 + `LoggerFunc`），一行接入 slog / zap / logrus
- Report 下载自动处理 `application/a-gzip` 内联 gzip 负载
- 截图上传封装 Apple 的 reservation → PUT → MD5 commit 三步流程，一次调用完成
- 全量 Context 支持（所有 API 都要求 `context.Context`，方便超时/取消）
- httptest + fixture 驱动的单测，覆盖率 > 85%

## 安装

```bash
go get github.com/godrealms/go-apple-sdk
```

需要 Go 1.23 及以上（使用了泛型和新版 `slices` 行为）。

## 目录

- [App Store Server API](#app-store-server-api)
  - [基础配置](#基础配置)
  - [测试服务器通知](#1-测试服务器通知)
  - [查询交易信息](#2-查询交易信息)
  - [查询订阅状态](#3-查询订阅状态)
  - [查询交易历史](#4-查询交易历史)
  - [发送消费信息](#5-发送消费信息)
- [App Store Connect API](#app-store-connect-api)
  - [初始化 Service](#初始化-service)
  - [Query Builder](#query-builder)
  - [错误处理](#错误处理)
  - [分页](#分页)
  - [结构化日志](#结构化日志)
  - [模块清单](#模块清单)
  - [常用场景示例](#常用场景示例)
- [示例程序](#示例程序)
- [贡献](#贡献)
- [许可证](#许可证)

---

## App Store Server API

### 基础配置

```go
import (
    Apple "github.com/godrealms/go-apple-sdk"
)

client := Apple.NewClient(
    true,                 // 是否为沙箱环境
    "YOUR_KEY_ID",        // 密钥 ID
    "YOUR_ISSUER_ID",     // Issuer ID
    "YOUR_BUNDLE_ID",     // 应用 Bundle ID
    "YOUR_PRIVATE_KEY",   // ES256 私钥 (PEM)
)
```

`client` 同时作为 App Store Server API 的入口和 App Store Connect API 的根 Client（通过 `client.AppStoreConnect()` 获取后者的 `*Service`）。

### 1. 测试服务器通知

```go
import AppStoreServer "github.com/godrealms/go-apple-sdk/app-store-server"

response, err := AppStoreServer.RequestTestNotification(client)
if err != nil {
    log.Fatal(err)
}
log.Println("TestNotificationToken:", response.TestNotificationToken)

testNotification, err := AppStoreServer.GetTestNotificationStatus(client, response.TestNotificationToken)
if err != nil {
    log.Fatal(err)
}
```

### 2. 查询交易信息

```go
info, err := AppStoreServer.GetTransactionInfo(client, "YOUR_TRANSACTION_ID")
if err != nil {
    log.Fatal(err)
}

transaction, err := info.SignedTransactionInfo.Decrypt()
if err != nil {
    log.Fatal(err)
}
```

### 3. 查询订阅状态

```go
subscriptions, err := AppStoreServer.GetAllSubscriptionStatuses(client, "TRANSACTION_ID")
if err != nil {
    log.Fatal(err)
}

for _, datum := range subscriptions.Data {
    for _, transaction := range datum.LastTransactions {
        renewalInfo, err := transaction.SignedRenewalInfo.Decrypt()
        if err != nil {
            log.Fatal(err)
        }
        transactionInfo, err := transaction.SignedTransactionInfo.Decrypt()
        if err != nil {
            log.Fatal(err)
        }
        _ = renewalInfo
        _ = transactionInfo
    }
}
```

### 4. 查询交易历史

```go
history, err := AppStoreServer.GetTransactionHistory(client, "TRANSACTION_ID", map[string]any{
    "sort": "DESCENDING",
    // "revision": "",
    // "startDate": "",
    // "endDate": "",
    // "productId": []string{},
    // "productType": []string{"AUTO_RENEWABLE", "NON_RENEWABLE"},
    // "inAppOwnershipType": []string{"FAMILY_SHARED", "PURCHASED"},
})
if err != nil {
    log.Fatal(err)
}

for _, transaction := range history.SignedTransactions {
    decrypted, err := transaction.Decrypt()
    if err != nil {
        log.Fatal(err)
    }
    _ = decrypted
}
```

### 5. 发送消费信息

```go
err := AppStoreServer.SendConsumptionInformation(client, "TRANSACTION_ID", &AppStoreServer.ConsumptionRequest{
    AccountTenure:            0,
    AppAccountToken:          "",
    ConsumptionStatus:        0,
    CustomerConsented:        true,
    DeliveryStatus:           0,
    LifetimeDollarsPurchased: 0,
    LifetimeDollarsRefunded:  0,
    Platform:                 1,
    PlayTime:                 0,
    SampleContentProvided:    false,
    UserStatus:               0,
})
```

---

## App Store Connect API

App Store Connect API 本身遵循 [JSON:API 规范](https://jsonapi.org/)，所有资源都是 `{type, id, attributes, relationships}` 的嵌套结构。本 SDK 在 `app-store-connect/` 下提供一套命名空间路由、强类型模型和 Builder，让调用者把注意力放在业务逻辑上。

### 初始化 Service

```go
import (
    Apple "github.com/godrealms/go-apple-sdk"
    AppStoreConnect "github.com/godrealms/go-apple-sdk/app-store-connect"
)

client := Apple.NewClient(false, keyID, issuerID, bundleID, privateKeyPEM)
svc := client.AppStoreConnect() // 类型为 *AppStoreConnect.Service
```

如果需要完全控制 `HTTPClient` / `BaseURL` / `UserAgent` / `Logger`，可以直接使用 `AppStoreConnect.New(Config)` 构造一个独立的 `Service`。

### Query Builder

```go
q := AppStoreConnect.NewQuery().
    Filter("bundleId", "com.acme.widgets").
    Fields("apps", "name", "bundleId", "sku").
    Include("appStoreVersions").
    Sort("-name").
    Limit(200)

resp, err := svc.Apps().List(ctx, q)
```

Builder 输出的 query string 按 key 字典序确定性排序，便于在 httptest 中断言请求 URL。

### 错误处理

```go
_, err := svc.Apps().Get(ctx, "bad-id", nil)

var apiErr *AppStoreConnect.APIError
if errors.As(err, &apiErr) {
    // Apple 明确拒绝的业务错误
    log.Printf("HTTP %d — %d errors", apiErr.StatusCode, len(apiErr.Errors))
    if apiErr.HasCode("STATE_ERROR") {
        // 处理特定业务错误码
    }
}

var clientErr *AppStoreConnect.ClientError
if errors.As(err, &clientErr) {
    // 本地构造 / 传输层故障（非 Apple 返回）
    log.Printf("local failure: %v", clientErr)
}
```

### 分页

所有 `List*` 方法都有同名的 `*Iterator` / `*ListAll` 变体：

```go
// 一次性全量消费（慎用，会把所有页加载进内存）
all, err := svc.Apps().ListAll(ctx, AppStoreConnect.NewQuery().Limit(200))

// 或者使用 Paginator 手动翻页
it := svc.Apps().ListIterator(AppStoreConnect.NewQuery().Limit(200))
for it.Next(ctx) {
    for _, app := range it.Page() {
        process(app)
    }
}
if err := it.Err(); err != nil {
    log.Fatal(err)
}
```

`Paginator` 自动跟随 Apple 返回的 `links.next`。

### 结构化日志

`Service` 支持一个可选的 `Logger` Hook，每个完成的 HTTP 往返会调用一次。接口故意保持最小，避免强绑定任何具体日志库：

```go
import "log/slog"

svc := AppStoreConnect.New(AppStoreConnect.Config{
    Authorizer: authorizer,
    Logger: AppStoreConnect.LoggerFunc(func(r AppStoreConnect.LogRecord) {
        level := slog.LevelInfo
        if r.Err != nil {
            level = slog.LevelError
        }
        slog.Log(context.Background(), level, "appstoreconnect",
            "method", r.Method,
            "url", r.URL,
            "status", r.StatusCode,
            "duration", r.Duration,
            "err", r.Err,
        )
    }),
})
```

`LogRecord` 不会暴露请求体 / 响应体；若需要更深入的埋点，请在 `Config.HTTPClient` 提供一个自定义的 `http.RoundTripper`。

### 模块清单

| 模块 | Go 入口 | 覆盖资源 | 典型用途 |
|---|---|---|---|
| Apps | `svc.Apps()` | `/v1/apps`, `/v1/apps/{id}` | 查询账号下应用、修改基础信息 |
| Reports | `svc.Reports()` | `/v1/salesReports`, `/v1/financeReports` | 自动下载销售 / 结算报表（gzip TSV） |
| CustomerReviews | `svc.CustomerReviews()` | `/v1/customerReviews`, `/v1/customerReviewResponses` | 拉取评分、回复用户 |
| Builds | `svc.Builds()` | `/v1/builds` | TestFlight 构建查询与更新 |
| BetaGroups | `svc.BetaGroups()` | `/v1/betaGroups` + relationships | 创建 TestFlight 组、添加测试人员和构建 |
| BetaTesterInvitations | `svc.BetaTesterInvitations()` | `/v1/betaTesterInvitations` | 触发测试邀请邮件 |
| BundleIDs | `svc.BundleIDs()` | `/v1/bundleIds` | 注册 / 删除 Bundle ID |
| Certificates | `svc.Certificates()` | `/v1/certificates` | base64 CSR 生成发布证书 |
| Profiles | `svc.Profiles()` | `/v1/profiles` | 制作 Provisioning Profile（自动绑定 bundleId / certs / devices） |
| Users | `svc.Users()` | `/v1/users` | 团队成员角色 / 可见性管理 |
| UserInvitations | `svc.UserInvitations()` | `/v1/userInvitations` | 邀请新成员，支持 `visibleApps` |
| AppStoreVersions | `svc.AppStoreVersions()` | `/v1/appStoreVersions` + `/relationships/build` | 版本 CRUD 与构建绑定 |
| AppStoreVersionSubmissions | `svc.AppStoreVersionSubmissions()` | `/v1/appStoreVersionSubmissions` | 提交 / 撤回审核 |
| AppStoreVersionLocalizations | `svc.AppStoreVersionLocalizations()` | `/v1/appStoreVersionLocalizations` | 多语言元数据（描述、关键词、Release Notes...） |
| AppScreenshotSets | `svc.AppScreenshotSets()` | `/v1/appScreenshotSets` | 按 `screenshotDisplayType` 分组 |
| AppScreenshots | `svc.AppScreenshots()` | `/v1/appScreenshots` + S3 upload | 一体化 `Upload()` 封装 Apple 的 reservation → PUT → MD5 commit |
| InAppPurchases | `svc.InAppPurchases()` | `/v1/inAppPurchasesV2` | 非订阅内购商品管理 |
| SubscriptionGroups | `svc.SubscriptionGroups()` | `/v1/subscriptionGroups` | 订阅组容器 |

### 常用场景示例

#### 自动化 App Store 提审

```go
// 1) 创建版本 → 2) 绑定构建 → 3) 写入 Release Notes → 4) 提交审核
ver, err := svc.AppStoreVersions().Create(ctx, AppStoreConnect.CreateAppStoreVersionRequest{
    AppID:         appID,
    Platform:      "IOS",
    VersionString: "1.2.3",
    ReleaseType:   "AFTER_APPROVAL",
})
if err != nil {
    return err
}

if err := svc.AppStoreVersions().SelectBuild(ctx, ver.Id, buildID); err != nil {
    return err
}

if _, err := svc.AppStoreVersionLocalizations().Create(ctx, AppStoreConnect.CreateAppStoreVersionLocalizationRequest{
    VersionID:       ver.Id,
    Locale:          "en-US",
    WhatsNew:        "- Faster spinner\n- Dark mode polish",
    PromotionalText: "Our fastest spinner yet!",
}); err != nil {
    return err
}

if _, err := svc.AppStoreVersionSubmissions().Submit(ctx, ver.Id); err != nil {
    return err
}
```

#### 一行上传截图

```go
png, _ := os.ReadFile("./hero-6.7.png")
shot, err := svc.AppScreenshots().Upload(ctx, screenshotSetID, "hero-6.7.png", png)
if err != nil {
    return err
}
log.Printf("screenshot %s delivered at %s", shot.Id, shot.Attributes.ImageAsset.TemplateURL)
```

`Upload` 内部会：

1. `POST /v1/appScreenshots` 进行 reservation
2. 按 Apple 返回的 `uploadOperations` 把字节切片 PUT 到上传端点（不带 JWT）
3. 计算 MD5 后 `PATCH /v1/appScreenshots/{id}` 标记 `uploaded=true` 并带上 `sourceFileChecksum`

#### TestFlight 分发

```go
bg, err := svc.BetaGroups().Create(ctx, AppStoreConnect.CreateBetaGroupRequest{
    AppID:             appID,
    Name:              "QA Team",
    PublicLinkEnabled: ptrBool(true),
})
if err != nil {
    return err
}
if err := svc.BetaGroups().AddBuilds(ctx, bg.Id, []string{buildID}); err != nil {
    return err
}
if _, err := svc.BetaTesterInvitations().Create(ctx, appID, testerID); err != nil {
    return err
}
```

#### 下载销售报表

```go
body, err := svc.Reports().DownloadSalesReport(ctx, AppStoreConnect.SalesReportRequest{
    VendorNumber: "1234567",
    ReportType:   "SALES",
    ReportSubType: "SUMMARY",
    Frequency:    "DAILY",
    ReportDate:   "2026-04-01",
})
if err != nil {
    return err
}
// body 已经是解 gzip 后的 TSV，可以直接 strings.Split / 解析
```

## 示例程序

所有可运行示例都在 `examples/` 下。每个子目录是一个独立的 `main` 包，填入自己的凭据即可编译运行：

### App Store Connect API

- `examples/app-store-connect/apps-list` — 列出账号下的所有应用
- `examples/app-store-connect/apps-update` — 更新应用基础信息
- `examples/app-store-connect/sales-report` — 下载销售报表
- `examples/app-store-connect/customer-reviews` — 读取 + 回复用户评价
- `examples/app-store-connect/testflight-distribute` — 创建 TestFlight 组并分发
- `examples/app-store-connect/provisioning-create` — 注册 Bundle ID + 签发证书 + 构建 Profile
- `examples/app-store-connect/users-invite` — 邀请团队成员
- `examples/app-store-connect/version-submit` — 创建版本、写 Release Notes、提交审核
- `examples/app-store-connect/screenshot-upload` — 一次调用完成 Apple 三步上传流程
- `examples/app-store-connect/iap-create` — 创建非订阅 IAP 和订阅组

### App Store Server API

- `examples/app-store-server/notifications-testing` — 请求 Apple 发送测试通知 + 查询状态
- `examples/app-store-server/transaction-info` — 查询单笔交易详情
- `examples/app-store-server/transaction-history` — 按订阅分组翻页拉取交易历史
- `examples/app-store-server/subscription-status` — 查询订阅状态
- `examples/app-store-server/consumption-info` — 上报消费信息配合退款决策
- `examples/app-store-server/order-lookup` — 根据订单号反查交易
- `examples/app-store-server/refund-history` — 拉取退款历史

### App Store Server Notifications v2

- `examples/app-store-server-notifications/v2` — 解析并校验 Apple 推送的 JWS 通知

## API 文档

完整的类型和方法文档请参考 [pkg.go.dev/github.com/godrealms/go-apple-sdk](https://pkg.go.dev/github.com/godrealms/go-apple-sdk)。

## 贡献

欢迎提交 Pull Request 和 Issue！在提交 PR 之前，请确保：

1. `go build ./...`、`go vet ./...`、`go test ./...` 全部通过
2. `golangci-lint run` 无新增告警（配置见 `.golangci.yml`）
3. 新增的 `app-store-connect` 功能附带 httptest + fixture 驱动的测试
4. 整个 `app-store-connect` 包的测试覆盖率保持在 80% 以上
5. 更新相关文档和示例

CI 会在 GitHub Actions 上跑 build / vet / lint / test + coverage gate，流程定义在 `.github/workflows/ci.yml`。

## 许可证

本项目采用 MIT 许可证。详情请参见 [LICENSE](LICENSE) 文件。

## 相关链接

- [App Store Server API 文档](https://developer.apple.com/documentation/appstoreserverapi)
- [App Store Server Notifications](https://developer.apple.com/documentation/appstoreservernotifications)
- [App Store Connect API 文档](https://developer.apple.com/documentation/appstoreconnectapi)
- [JSON:API 规范](https://jsonapi.org/)
