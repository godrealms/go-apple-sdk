# Go Apple SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/godrealms/go-apple-sdk.svg)](https://pkg.go.dev/github.com/godrealms/go-apple-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/godrealms/go-apple-sdk)](https://goreportcard.com/report/github.com/godrealms/go-apple-sdk)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Go Apple SDK 是一个用于与 Apple App Store Server API 和 App Store Connect API 交互的 Go 语言 SDK。该 SDK 提供了简单易用的接口来处理应用内购买、订阅和 App Store 通知等功能。

## 功能特性

- **App Store Server API 支持**
- 交易信息查询
- 订阅状态查询
- 消费信息查询
- 订单查询
- 退款查询
- 应用内购买历史记录
- 服务器通知测试
- **类型安全的 API 设计**
- **完整的错误处理**
- **详细的使用示例**

## 安装

```bash
go get github.com/godrealms/go-apple-sdk
```

## 快速开始

### 基础配置

```go
import (
  Apple "github.com/godrealms/go-apple-sdk"
)

// 创建客户端实例
client := Apple.NewClient(
  true,                // 是否为沙箱环境
  "YOUR_KEY_ID",      // 您的密钥 ID
  "YOUR_ISSUER_ID",   // 您的发行者 ID
  "YOUR_BUNDLE_ID",   // 您的应用 Bundle ID
  "YOUR_PRIVATE_KEY", // 您的私钥
)
```

### 示例用法

#### 1. 测试服务器通知

```go
import AppStoreServer "github.com/godrealms/go-apple-sdk/app-store-server"

// 请求测试通知
response, err := AppStoreServer.RequestTestNotification(client)
if err != nil {
  log.Fatal(err)
}
log.Println("TestNotificationToken:", response.TestNotificationToken)

// 获取测试通知状态
testNotification, err := AppStoreServer.GetTestNotificationStatus(client, response.TestNotificationToken)
if err != nil {
  log.Fatal(err)
}
```

#### 2. 查询交易信息

```go
info, err := AppStoreServer.GetTransactionInfo(client, "YOUR_TRANSACTION_ID")
if err != nil {
  log.Fatal(err)
}

// 解密交易信息
transaction, err := info.SignedTransactionInfo.Decrypt()
if err != nil {
  log.Fatal(err)
}
```

#### 3. 查询订阅状态

```go
subscriptions, err := AppStoreServer.GetAllSubscriptionStatuses(client, "TRANSACTION_ID")
if err != nil {
  log.Fatal(err)
}

// 遍历订阅信息
for _, datum := range subscriptions.Data {
  for _, transaction := range datum.LastTransactions {
      // 解密续订信息
      renewalInfo, err := transaction.SignedRenewalInfo.Decrypt()
      if err != nil {
          log.Fatal(err)
      }
      
      // 解密交易信息
      transactionInfo, err := transaction.SignedTransactionInfo.Decrypt()
      if err != nil {
          log.Fatal(err)
      }
  }
}
```

#### 4. 查询交易历史

```go
history, err := AppStoreServer.GetTransactionHistory(client, "TRANSACTION_ID", map[string]any{
    // 可选参数 
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

// 解密交易记录
for _, transaction := range history.SignedTransactions {
  decrypted, err := transaction.Decrypt()
  if err != nil {
      log.Fatal(err)
  }
}
```

#### 5. 发送消费信息

```go
err := AppStoreServer.SendConsumptionInformation(client, "TRANSACTION_ID", &AppStoreServer.ConsumptionRequest{
  AccountTenure: 0,
  AppAccountToken: "",
  ConsumptionStatus: 0,
  CustomerConsented: true,
  DeliveryStatus: 0,
  LifetimeDollarsPurchased: 0,
  LifetimeDollarsRefunded: 0,
  Platform: 1,
  PlayTime: 0,
  SampleContentProvided: false,
  UserStatus: 0,
})
```

## API 文档

完整的 API 文档请参考 [GoDoc](https://pkg.go.dev/github.com/godrealms/go-apple-sdk)。

## 贡献

欢迎提交 Pull Request 和 Issue！在提交 PR 之前，请确保：

1. 代码通过所有测试
2. 新功能包含相应的测试用例
3. 更新相关文档

## 许可证

本项目采用 MIT 许可证。详情请参见 [LICENSE](LICENSE) 文件。

## 相关链接

- [App Store Server API 文档](https://developer.apple.com/documentation/appstoreserverapi)
- [App Store Server Notifications](https://developer.apple.com/documentation/appstoreservernotifications)
- [App Store Connect API](https://developer.apple.com/documentation/appstoreconnectapi)