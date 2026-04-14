# 🚀 go-apple-sdk: App Store Connect API 完整实现与演进计划

## 背景与愿景 (Background & Vision)

目前 `go-apple-sdk` 已经实现了大部分 App Store Server API 的功能（如交易查询、历史记录、退款、订阅管理和服务器通知）。然而，为了让本 SDK 成为一个名副其实且全面的 Apple 生态集成工具，我们需要完整支持 **App Store Connect API**。

App Store Connect API 是一个极其庞大且复杂的 REST API 系统，包含了数百个 Endpoint，用于自动化执行在 Apple Developer 网站和 App Store Connect 上的几乎所有任务。它完全遵循 [JSON:API 规范](https://jsonapi.org/)。

由于其体量和复杂性，我们必须制定一个结构化、分阶段的实施计划，确保架构的高可扩展性、类型安全和易用性。

---

## 🏗️ 核心架构挑战与设计原则 (Architecture & Design Principles)

1. **JSON:API 规范的深度解析**:
   * App Store Connect API 的请求和响应结构非常复杂，大量使用 `data`, `attributes`, `relationships`, `included` 等嵌套层级。
   * **解决方案**: 在 `types/` 目录下设计高度复用且支持泛型的 Go 结构体（如 `Document[T]`, `ResourceObject`, `Relationship`, `Included`）来反序列化这些数据，避免重复定义。

2. **复杂的查询参数 (Query Builder)**:
   * 接口广泛使用结构化参数，如过滤 (`filter[bundleId]=...`)、字段选择 (`fields[apps]=...`)、关联包含 (`include=...`)、排序 (`sort=-name`) 和分页 (`limit=200`)。
   * **解决方案**: 实现一个强类型的流式 Query Builder（例如 `NewQuery().Filter("bundleId", "com.example.app").Include("appStoreVersions").Limit(200)`），自动转换为正确的 URL 编码。

3. **分页 (Pagination) 与速率限制 (Rate Limiting)**:
   * API 响应中包含 `links.next` 用于游标分页。同时，Apple 对 API 有严格的调用频率限制。
   * **解决方案**: 封装自动处理游标分页的迭代器 (Iterator) 模式。在 HTTP Client 层面增加遵循 `X-Rate-Limit` 响应头的自动退避重试 (Backoff & Retry) 机制。

4. **标准化错误处理 (Error Handling)**:
   * App Store Connect 返回的错误通常是一个包含多个错误对象的数组 (`{"errors": [{"status": "409", "code": "STATE_ERROR", "title": "..."}]}`)。
   * **解决方案**: 定义专门的 `AppStoreConnectError` 类型，实现 `error` 接口，方便调用者通过 `errors.As` 提取具体业务错误码。

5. **高内聚低耦合的模块化路由**:
   * 不能把数百个方法都挂载到单一的 `*Apple.Client` 实例上。
   * **解决方案**: 采用命名空间路由设计。例如：
     ```go
     client.AppStoreConnect().Apps().List(ctx, query)
     client.AppStoreConnect().TestFlight().GetBuild(ctx, buildId)
     client.AppStoreConnect().Sales().DownloadReport(ctx, filter)
     ```

---

## 📅 实施路线图 (Roadmap)

### 阶段一：基础设施建设 (Foundation)
**目标**：搭建坚实的底层骨架，解决现有工程问题，确保后续添加接口时只需关注业务逻辑。

- [x] **修复现有工程问题**: 重构 `examples/` 目录结构，解决多个 `main` 函数冲突导致的编译和测试失败问题。每个示例拆到独立子目录 (`examples/app-store-server/<feature>/main.go`)，`go build ./...` 与 `go vet ./...` 全量通过。
- [x] **JSON:API 基础模型**: 在 `app-store-connect/jsonapi.go` 实现泛型 `Document[T]`、`Resource[T]`、`Relationship` (含 `AsOne`/`AsMany`)、`ResourceIdentifier`、`Links`。支持单资源与集合两种文档形态，未知字段容忍。
- [x] **Query Builder**: `app-store-connect/query.go` 中的 `Query` 支持 `Filter` / `Fields` / `Include` / `Sort` / `Limit` / `Cursor` / `Set` / `Clone`，编码结果按 key 字母序确定性排序，便于测试断言。
- [x] **错误处理与分页模块**: `errors.go` 中的 `APIError` 实现 `error` 接口并通过 `errors.As` 提取、提供 `HasCode` 快捷方法；`ClientError` 区分本地故障。`paginator.go` 中的泛型 `Paginator[T]` 自动跟随 `links.next`，并提供 `All()` 一次性消费。
- [x] **测试脚手架**: `testing_test.go` 提供 `newTestService` + `loadFixture` 帮助函数，基于 `net/http/httptest.Server` + `testdata/*.json` 夹具驱动。覆盖率 80.7%，达到阶段五目标线。
- [x] **跑通首个接口**: `apps.go` 实现 `AppsService.List`、`AppsService.Get`、`AppsService.ListIterator`、`AppsService.ListAll`。根 `*Apple.Client` 通过 `Client.AppStoreConnect()` 工厂方法注入无 scope JWT 的 `Authorizer`，打通鉴权 → Query Builder → JSON:API 解码 → 错误处理 → 分页 全链路。示例见 `examples/app-store-connect/apps-list/main.go`。

### 阶段二：核心高频业务落地 (Core Modules)
**目标**：优先实现开发者和数据团队最关心的三大核心痛点功能。

- [ ] **App 基础信息管理 (Apps & App Infos)**
  - 获取账号下应用列表及单应用详情 (`GET /v1/apps`, `GET /v1/apps/{id}`)
  - 修改应用基础信息 (`PATCH /v1/apps/{id}`)
- [ ] **销售与财务报表下载 (Sales & Finance Reports)**
  - 下载每日/每周/每月销售报表，自动处理 gzip 解压 (`GET /v1/salesReports`)
  - 下载财务结算报表 (`GET /v1/financeReports`)
- [ ] **用户评论管理 (Customer Reviews)**
  - 拉取 App Store 用户评价 (`GET /v1/apps/{id}/customerReviews`)
  - 自动化回复用户评价 (`POST /v1/customerReviewResponses`)

### 阶段三：CI/CD 与 TestFlight 自动化 (DevOps Automation)
**目标**：赋能开发团队，支持将 SDK 集成到 CI/CD 流水线（如 GitHub Actions, Jenkins）中。

- [ ] **TestFlight 构建与测试管理**
  - 获取和管理构建版本 (`GET /v1/builds`, `PATCH /v1/builds/{id}`)
  - 管理测试群组 (`GET/POST /v1/betaGroups`)
  - 邀请内部/外部测试人员 (`POST /v1/betaTesterInvitations`)
- [ ] **证书与描述文件管理 (Provisioning)**
  - 管理 Bundle ID (`GET/POST /v1/bundleIds`)
  - 查询和管理证书 (`GET /v1/certificates`)
  - 管理 Provisioning Profiles (`GET/POST /v1/profiles`)
- [ ] **用户与权限管理 (Users & Roles)**
  - 查询团队成员及权限 (`GET /v1/users`)
  - 发送团队邀请 (`POST /v1/userInvitations`)

### 阶段四：App Store 提审与元数据自动化 (Store Submission)
**目标**：实现类似 `fastlane deliver` 的全自动 App Store 提审流和商品管理。

- [ ] **版本与提审管理 (App Store Versions & Reviews)**
  - 创建和配置新的 App Store 版本 (`POST /v1/appStoreVersions`)
  - 提交应用审核 (`POST /v1/appStoreVersionSubmissions`)
- [ ] **应用商店元数据与资产 (Metadata & Media)**
  - 更新多语言的推广文本和 Release Notes (`PATCH /v1/appStoreVersionLocalizations/{id}`)
  - 上传应用截图和预览视频 (`POST /v1/appScreenshotSets`, `POST /v1/appScreenshots`)
- [ ] **应用内购买配置管理 (In-App Purchases Configuration)**
  - 动态创建和修改内购商品 (`POST /v1/inAppPurchasesV2`)
  - 管理订阅组和价格档位配置 (`POST /v1/subscriptionGroups`)

### 阶段五：工程化规范与社区生态 (Engineering & OSS)
**目标**：打造生产级可用、文档详尽、易于开源社区贡献的高质量 Go 项目。

- [ ] **完善使用文档 (Documentation)**
  - 编写包含丰富示例的 `README.md` 和 GoDoc 注释。
  - 为 App Store Connect API 的不同模块提供详细的使用教程。
- [ ] **CI 流水线建设 (GitHub Actions)**
  - 配置 `golangci-lint` 保证代码风格一致性（如强制 `snake_case` 文件命名）。
  - 自动运行单元测试并统计代码覆盖率（目标覆盖率 > 80%）。
- [ ] **健壮性与可观测性**
  - 增加结构化的日志输出选项。
  - 提供 Context 支持，允许用户控制请求超时和取消 (`req.SetContext(ctx)`)。

---

## 🚀 行动指南 (Next Steps)

为了稳步推进本计划，建议严格按照阶段顺序执行。**当前的当务之急是启动「阶段一」**：

1. **清理历史包袱**: 修复 `examples/` 目录的编译错误。
2. **基建代码提交**: 在 `types/` 下建立 JSON:API 相关的基础结构体。
3. **完成探路接口**: 在 `app-store-connect/` 目录下完成 `Apps` 服务的骨架并实现 `GET /v1/apps`。

*(本文档将随着开发的推进进行持续更新和状态追踪。)*