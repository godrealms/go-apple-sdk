# go-apple-sdk：App Store Connect API 全量覆盖设计方案

- **日期**：2026-04-16
- **状态**：Draft，待实施
- **前置阶段**：Phase 1 - 5（已完成，18 个手写 Service + 基础设施）
- **本阶段编号**：Phase 6（子阶段 6.1 - 6.6）

---

## 1. 背景与目标

### 1.1 现状

`go-apple-sdk` 当前对 App Store Connect API 的覆盖率约为 **20%**（18 个手写 Service，~80 个 endpoint）。Apple 官方 OpenAPI 规范包含 **70+ 资源类型、400+ endpoint**。覆盖面与"完整支持 Apple 开发者生态"的仓库愿景尚有明显差距。

### 1.2 目标

| 维度 | 目标 |
|---|---|
| **API 覆盖率** | 100%——覆盖 Apple OpenAPI 规范中的所有 resource 类型和 endpoint |
| **用户体验** | 单一 `*Service`、统一 API 表面；用户感知不到"手写 vs 生成"的区别 |
| **可维护性** | Apple 发布新 spec → `make update-spec` 一键同步，diff 即 changelog |
| **复现性** | Spec SHA256 锁定；任何时间跑 gen 结果字节级一致 |
| **零工具链依赖** | 用户 `go get` 后立即可用，不需要运行任何代码生成器 |
| **向后兼容** | 现有 18 个手写 Service 的 API 契约零改动 |

### 1.3 非目标

- **不**为 gen 代码提供 Builder Pattern（Builder 仅属于手写精品 Service）
- **不**生成 streaming / WebSocket 客户端（Apple API 无此类 endpoint）
- **不**在 gen 过程中做业务规则校验（交给 Apple 服务端拒绝非法输入）
- **不**尝试"智能合并" spec 差异——每次 spec 更新都是全量重跑

---

## 2. 核心决策

以下五项在本次 brainstorming 中已经与用户确认，作为整个设计的基石。

| # | 决策点 | 选定方案 |
|---|---|---|
| 1 | **执行策略** | A + B 混合：已有 18 个手写精品保留；剩余 50+ 资源用 codegen 生成 |
| 2 | **包共存模式** | Z 方案——单一 API 表面 + Skip list。用户只看到 `app-store-connect/` 一套 Service，感知不到哪些是手写 / 哪些是生成 |
| 3 | **测试策略** | 三层：手写 Service 继续手写测；gen 方法每个自动产出 smoke test；整份 gen 出一份 golden 清单做 round-trip freshness 校验 |
| 4 | **OpenAPI spec 来源** | Vendor 进仓库 + SHA256 锁定。Spec 本身 commit 到 `internal/cmd/gen-asc/spec/`，diff 即 Apple API changelog |
| 5 | **生成器调用方式** | i + ii 并存：`internal/cmd/gen-asc/` 是可独立运行的 Go 程序（产物 commit 入库）+ `//go:generate` 指令 + Makefile `gen` / `update-spec` 目标 |

---

## 3. 架构

### 3.1 仓库布局

```
go-apple-sdk/
├── app-store-connect/                    # 用户 import 的公共包（唯一入口）
│   ├── service.go                        # [手写] Service 核心，do/doRaw/Logger
│   ├── jsonapi.go                        # [手写] Document[T] / Resource[T] 基础设施
│   ├── query.go                          # [手写] Query Builder
│   ├── paginator.go                      # [手写] 泛型分页
│   ├── errors.go                         # [手写] APIError / ClientError
│   ├── auth.go                           # [手写] ES256 JWT
│   │
│   ├── apps.go                           # [手写精品] 18 个现有 Service
│   ├── builds.go                         # [手写精品]
│   ├── ... (其它 16 个手写精品)
│   │
│   ├── zz_generated_types.go             # [gen] 所有 gen Service 的 attributes / relationships 类型
│   ├── zz_generated_service_wiring.go    # [gen] Service struct 追加字段 + New() 注入 + 访问器
│   ├── zz_generated_analytics_reports.go # [gen] 一个资源一个文件
│   ├── zz_generated_app_prices.go        # [gen]
│   ├── ... (约 50+ gen Service 文件)
│   │
│   ├── zz_generated_analytics_reports_test.go    # [gen] smoke test，一个方法一条
│   ├── zz_generated_app_prices_test.go           # [gen]
│   ├── ... (和 gen Service 文件一一对应)
│   │
│   ├── zz_generated_golden.json          # [gen] 黄金清单快照
│   ├── golden_test.go                    # [手写] 读 golden.json 做 round-trip / freshness 校验
│   │
│   └── testdata/                         # 现有手写 fixture 保持不变
│
├── internal/cmd/gen-asc/                 # 生成器（internal/ 防止外部 import）
│   ├── main.go                           # CLI entrypoint
│   ├── parser/                           # OpenAPI schema → IR
│   │   └── parser.go
│   ├── ir/                               # 内部表达
│   │   └── ir.go
│   ├── render/                           # text/template 渲染
│   │   ├── render.go
│   │   └── templates/
│   │       ├── service.go.tmpl
│   │       ├── smoke_test.go.tmpl
│   │       ├── types.go.tmpl
│   │       ├── service_wiring.go.tmpl
│   │       └── golden.json.tmpl
│   ├── skip/
│   │   ├── skip.go
│   │   └── skip.yaml                     # 手写精品 Service 已覆盖的 OpenAPI 资源名清单
│   ├── naming/
│   │   ├── naming.go                     # camelCase ↔ Go PascalCase 转换
│   │   └── acronyms.go                   # URL / ID / API / JWT / JSON 等保留词
│   └── spec/
│       ├── app_store_connect_api_openapi.json   # Vendor 的 OpenAPI spec
│       └── version.go                    # SpecVersion / SpecSHA256 / SpecSource 常量
│
├── scripts/
│   └── update-spec.sh                    # 下载 ZIP → 解压 → 校验 SHA → 替换 vendor
│
├── Makefile                              # gen / update-spec / test / lint 目标
│
└── docs/superpowers/specs/
    └── 2026-04-16-asc-full-coverage-design.md    # 本文档
```

### 3.2 数据流

```
Apple OpenAPI spec (JSON, vendor 在 internal/cmd/gen-asc/spec/)
    │
    ▼
parser/      encoding/json → map → normalize
    │
    ▼
ir/          []Resource / []Endpoint / []Type / Metadata
    │
    ▼
skip/        根据 skip.yaml 过滤掉手写精品资源
    │
    ▼
render/      text/template + gofmt
    │
    ▼
app-store-connect/zz_generated_*.go (commit 入库)
app-store-connect/zz_generated_golden.json (commit 入库)
```

### 3.3 用户视角的 API

```go
import asc "github.com/godrealms/go-apple-sdk/app-store-connect"

svc := asc.New(asc.Config{
    Authorizer: auth,
    Logger:     logger,
})

// —— 手写精品（带 Builder、带校验、带特殊 flow）——
svc.Apps().Update(ctx, appID,
    asc.NewAppUpdate().PrimaryLocale("zh-CN"))

svc.AppScreenshots().Upload(ctx, setID, "screen.png", imageBytes)

// —— 生成代码（直白 JSON:API CRUD，用 *Query 控制 filter/include/fields）——
reports, err := svc.AnalyticsReports().List(ctx,
    asc.NewQuery().Filter("category", "APP_USAGE").Include("instances"))

price, err := svc.AppPricePoints().Get(ctx, pricePointID, nil)

// 两者共享同一套 *Document[T] / *APIError / Paginator / Logger
```

用户**看不到**哪些是手写哪些是 gen。

---

## 4. 生成器内部设计

### 4.1 IR（Intermediate Representation）

把 OpenAPI 的嵌套 schema 压平成 Go 友好的结构。所有模板只对 IR 打洞，彻底避免模板里处理 OpenAPI 细节。

```go
// internal/cmd/gen-asc/ir/ir.go
type Resource struct {
    Name       string        // "AnalyticsReports"（Go 命名）
    ApiName    string        // "analyticsReports"（JSON:API type）
    Operations []Operation
    Attrs      []Field       // attributes 对象字段
    Rels       []Relationship
    DocURL     string
}

type Operation struct {
    Name         string   // "List" / "Get" / "Create" / "Update" / "Delete" / "ListForApp" ...
    HTTPMethod   string   // "GET" / "POST" / ...
    PathTemplate string   // "/v1/analyticsReports/{id}"
    PathParams   []Field
    QueryParams  []Field  // filter/include/fields 白名单
    RequestBody  *Type
    ResponseBody *Type
    DocURL       string
    Deprecated   bool
}

type Field struct {
    Name     string
    JsonName string
    GoType   string
    Required bool
    Comment  string
}

type Type struct {
    Name   string
    Fields []Field
}
```

### 4.2 OpenAPI → Go 转换规则

| OpenAPI 元素 | IR 处理 | 最终 Go 形态 |
|---|---|---|
| `components.schemas.AnalyticsReport` | Resource | `AnalyticsReport` struct + `AnalyticsReportsService` |
| `paths./v1/analyticsReports.get` | Operation(List) | `(s *AnalyticsReportsService) List(ctx, query) (*Document[[]AnalyticsReport], error)` |
| `paths./v1/analyticsReports/{id}.get` | Operation(Get) | `Get(ctx, id, query) (*Document[AnalyticsReport], error)` |
| `string` (format: date-time) | Field | `time.Time` |
| `string` (enum: [...]) | Field | `type AnalyticsReportCategory string` + const block |
| `array: {items: $ref}` | Field | `[]RefType` |
| `required: [...]` | Field.Required | `omitempty` 是否写入 JSON tag |
| `nullable: true` | Field | `*string` / `*int` |
| 4xx / 5xx response schema | 忽略 | 统一走 `parseErrorBody` → `*APIError` |
| `relationships.app` | Relationship | `AppRel Relationship` 字段 |

### 4.3 命名规约

| 场景 | 规则 |
|---|---|
| Resource 复数 → Service | `analyticsReports` → `AnalyticsReportsService` |
| Resource 单数 → 类型 | `analyticsReport` → `AnalyticsReport` |
| 缩写保留 | `URL` / `ID` / `API` / `JWT` / `JSON` 整体大写 |
| JSON 字段 → Go 字段 | `bundleId` → `BundleID`，`purchaseUrl` → `PurchaseURL` |
| Go 关键字冲突 | `type` → 加前缀 `ResourceType`，不用 `Type_` |
| Path 参数 | `{id}` → 方法签名里的 `id string` |
| Attributes 子类型冲突 | 一律加资源前缀：`AnalyticsReportAttributes` 而非裸 `Attributes` |

`internal/cmd/gen-asc/naming/acronyms.go` 作为权威缩写表，生成器每次加载。

### 4.4 模板样例

#### `service.go.tmpl`（简化）

```go
// Code generated by internal/cmd/gen-asc. DO NOT EDIT.
// Source: {{.SpecSource}} (v{{.SpecVersion}}, sha256={{.SpecSHA256}})

package AppStoreConnect

import "context"

// {{.Name}}Service exposes the "{{.ApiName}}" resource on App Store Connect.
//
// See {{.DocURL}}
type {{.Name}}Service struct {
    svc *Service
}

{{range .Operations}}
// {{.Name}} {{.Comment}}
//
// {{.HTTPMethod}} {{.PathTemplate}}
{{if .Deprecated}}// Deprecated: {{.DeprecationNote}}{{end}}
func (s *{{$.Name}}Service) {{.Name}}(
    ctx context.Context,
    {{range .PathParams}}{{.Name}} {{.GoType}}, {{end}}
    {{if .HasBody}}body *{{.RequestBody.Name}}, {{end}}
    query *Query,
) (*Document[{{.ResponseBody.Name}}], error) {
    path := {{.PathBuilder}}
    out := &Document[{{.ResponseBody.Name}}]{}
    _, err := s.svc.do(ctx, "{{.HTTPMethod}}", path, query,
        {{if .HasBody}}body{{else}}nil{{end}}, out)
    return out, err
}
{{end}}
```

#### `smoke_test.go.tmpl`（简化）

```go
// Code generated by internal/cmd/gen-asc. DO NOT EDIT.

package AppStoreConnect

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

{{range .Operations}}
func Test{{$.Name}}Service_{{.Name}}_Generated(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if got, want := r.Method, "{{.HTTPMethod}}"; got != want {
            t.Errorf("method = %q, want %q", got, want)
        }
        // {{.PathAssertion}}
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(`{{.ExampleResponseBody}}`))
    }))
    defer srv.Close()

    svc := New(Config{BaseURL: srv.URL, Authorizer: noopAuthorizer})
    {{.CallSite}}
    if err != nil {
        t.Fatalf("{{.Name}}: %v", err)
    }
}
{{end}}
```

### 4.5 Skip List 格式

`internal/cmd/gen-asc/skip/skip.yaml`：

> **说明**：skip 列表的条目数（当前 20）**不等于**手写 Service 的个数（18）。少数手写 Service 覆盖了一个以上的 OpenAPI 资源（例如 `CustomerReviewsService` 同时处理 `customerReviews` 和 `customerReviewResponses`；`ReportsService` 同时处理 `salesReports` 和 `financeReports`）。skip.yaml 以 **OpenAPI 资源名** 为粒度，确保 generator 不会漏过任何一个被手写 Service 接管的资源。

```yaml
# 这些资源由手写精品 Service 覆盖，generator 跳过不生成。
# 新增手写 Service 时同步追加到这里；删除手写 Service 时从这里移除。
# Key 是 OpenAPI 里的 resource type（lowerCamel），不是 Go 命名。
# 条目粒度 = OpenAPI 资源；一个手写 Service 可以占多行。

skip:
  - apps
  - appStoreVersions
  - appStoreVersionSubmissions
  - appStoreVersionLocalizations
  - appScreenshotSets
  - appScreenshots
  - builds
  - betaGroups
  - betaTesterInvitations
  - bundleIds
  - certificates
  - profiles
  - users
  - userInvitations
  - customerReviews
  - customerReviewResponses
  - salesReports
  - financeReports
  - inAppPurchases
  - subscriptionGroups
```

Generator 启动时加载 YAML，过滤 IR 里的 Resource：跳过的不生成 `.go`、不生成 `_test.go`、不进 golden 清单、类型也不在 `zz_generated_types.go` 里重复定义。

### 4.6 Golden 清单格式

`app-store-connect/zz_generated_golden.json`：

```json
{
  "spec_version": "3.7.1",
  "spec_sha256": "7a2b...",
  "generated_at": "2026-04-16T00:00:00Z",
  "services": {
    "AnalyticsReportsService": {
      "api_name": "analyticsReports",
      "operations": [
        {"name": "List",   "method": "GET",    "path": "/v1/analyticsReports"},
        {"name": "Get",    "method": "GET",    "path": "/v1/analyticsReports/{id}"},
        {"name": "Create", "method": "POST",   "path": "/v1/analyticsReports"},
        {"name": "Delete", "method": "DELETE", "path": "/v1/analyticsReports/{id}"}
      ]
    }
  }
}
```

Apple 更新 spec 后跑 gen → `git diff zz_generated_golden.json` 一眼看全部变更，不用读几百个 `.go` 文件 diff。

### 4.7 Spec 版本锁定

`internal/cmd/gen-asc/spec/version.go`：

```go
package spec

// 这些常量必须和 vendor 的 OpenAPI 文件保持同步。
// 更新 spec 文件时必须同时更新这里的 SpecVersion 和 SpecSHA256。
// 生成器启动时会校验 vendor 文件的 SHA256，不匹配则 fail-fast。

const (
    SpecVersion = "3.7.1"
    SpecSHA256  = "7a2b..."  // sha256(vendor/app_store_connect_api_openapi.json)
    SpecSource  = "https://developer.apple.com/sample-code/app-store-connect/app-store-connect-openapi-specification.zip"
    SpecDate    = "2026-03-15"
)
```

---

## 5. 测试策略

### 5.1 三层测试

#### 第一层：手写精品 Service（现有 18 个）

- 位置：`app-store-connect/apps_test.go` 等（已存在）
- 风格：`httptest` + fixture + 业务 round-trip
- 门槛：每个手写 Service 覆盖率 ≥ 80%
- 变更：本阶段保持原样

#### 第二层：Gen Service smoke test

- 位置：`app-store-connect/zz_generated_*_test.go`
- 风格：每个 gen 方法自动产出一个 `Test{Service}_{Method}_Generated`
- 断言三件事：
  - **URL 正确**：method、path、path 参数替换、query 序列化
  - **请求体正确**：POST/PATCH 的 JSON payload 能 marshal 且结构匹配
  - **响应体正确**：mock 示例 JSON 能 unmarshal 到目标类型、`Document[T].Data.Id` 非空
- mock 响应来源（按优先级）：
  1. OpenAPI `examples` 字段
  2. OpenAPI `example` 字段
  3. 按 schema 合成最小 valid JSON（string→`"x"`, int→`1`, bool→`false`, array→`[]`, object→递归）
- 门槛：所有 gen 方法 100% 有一个 smoke test（由模板保证）

#### 第三层：Golden 清单 round-trip（一个，手写）

- 位置：`app-store-connect/golden_test.go`（**手写**，不生成）
- 职责：
  1. 加载 `zz_generated_golden.json`
  2. 反射枚举所有 `*Service` 方法
  3. 断言每个 gen 方法都在 golden 里有对应记录（防漏生成/漏提交）
- 附加：`TestGoldenFreshness` 比对 `SpecSHA256` 常量和 vendor 文件实际 SHA256，检测"spec 换了但忘记跑 gen"

### 5.2 覆盖率目标

| 范围 | 目标 |
|---|---|
| `app-store-connect/` 手写文件 | ≥ 80%（保持现状） |
| `app-store-connect/zz_generated_*.go` | ≥ 60%（smoke test 只验通路） |
| 全包合计 | ≥ 70% |

---

## 6. 滚动计划（6 个子阶段）

| 阶段 | 目标 | 主要产物 | 备注 |
|---|---|---|---|
| **6.1** Generator 脚手架 | `internal/cmd/gen-asc/` 可运行，能解析 spec、输出 IR 的 JSON dump（不生成代码） | `parser/` `ir/` `main.go` `spec/*.json` `version.go` `skip/skip.yaml`(空列表) | 最复杂，单独一个阶段 |
| **6.2** 类型生成 | 从 IR 生成 `zz_generated_types.go`（所有 gen 资源的 attributes / relationships 类型） | 模板 `types.go.tmpl` + render | 模板打磨重点 |
| **6.3** Service 生成 | 每个非 skip 资源生成 `zz_generated_<resource>.go` + wiring | 模板 `service.go.tmpl` `service_wiring.go.tmpl` | 整合现有 `Service` struct 追加字段 |
| **6.4** Smoke test + Golden | 每个 gen 方法一个 smoke test + 一份 golden 快照 + round-trip test | 模板 `smoke_test.go.tmpl` + `golden_test.go`(手写) | |
| **6.5** 首次全量 gen + Skip list 落地 | 跑一次完整生成，填充 skip list，解决命名/类型冲突直到编译通过 | 第一个可 build 的全量 gen 产物 | 大概率 build 失败→迭代修 |
| **6.6** Makefile / go:generate / 文档 | `make gen` / `make update-spec` / `//go:generate` / README / PLAN 更新 | Makefile / scripts/update-spec.sh / 文档 | 收尾 |

每个阶段按既有节奏：实现 → 测试 → commit → 下一阶段。

---

## 7. 现有 18 个手写 Service 的迁移

**关键原则：不动它们。**

- Skip list 保证 generator 不碰它们
- 它们的 API 契约、文件路径、测试、Builder 全部保持现状
- 用户侧**零破坏性变更**

唯一需要协调的是**类型重复**问题：

1. 手写 `apps.go` 已定义 `App` / `AppAttributes` 等类型
2. `zz_generated_types.go` 生成时同样会跳过 skip list 里的资源，**不重复定义**
3. 其它 gen Service 如果 response 里引用 `App`（例如 `AppPrices` 的 relationship），import 的就是手写的那个 `App`——**无缝对接**

换句话说：skip list 既跳过了 Service 生成，也跳过了 Type 生成——一次过滤，两层保护。

---

## 8. 风险与对策

| 风险 | 影响 | 缓解 |
|---|---|---|
| **OpenAPI spec 不完整**——缺 example、缺 description、甚至缺 endpoint | Smoke test 无数据、类型缺字段 | 先跑 IR dump 看缺口；fallback 合成策略；遇到 spec bug 把该资源临时加进 skip list |
| **类型命名冲突**——两个资源都有 `Attributes` 子类型 | Go 编译失败 | Attributes 一律带资源前缀 |
| **Path 模板参数顺序** | 方法调用错 | IR 里 PathParams 按 spec 出现顺序存，生成时保持同序 |
| **模板打磨需要多轮 gen-diff-fix** | Phase 6.5 迭代次数多 | 生成器支持 `-dry-run` 和 `-only=AnalyticsReports` 增量生成 |
| **gen 文件太多冲击 `go build` 速度** | 构建变慢 10-30% | 可接受；go 编译器对几百个短文件没问题 |
| **手写 Service 和 gen Service 类型交叉依赖变复杂** | 改手写可能破坏 gen | 每次 `make gen` 后跑全量测试作为闸门；golden freshness test 兜底 |
| **Apple spec 未来引入 breaking change** | 用户代码被破坏 | 遵守 Go semver：不兼容变更 → v2.0.0；向下兼容 → v1.x |
| **Generator 自身有 bug 产出错代码** | 用户拿到错误 Service | Smoke test 是第一道墙；golden round-trip 是第二道墙 |

---

## 9. 成功标准

Phase 6.1 - 6.6 完成时必须全部满足以下条件（验收清单，未勾选项表示尚未验收）：

- [ ] `svc.XxxService()` 方法数从 18 增加到 70+
- [ ] Apple OpenAPI spec 里的所有非 skip 资源 100% 有对应 Service
- [ ] 现有 18 个手写 Service 的 API 契约零改动
- [ ] `go test -race ./...` 全绿，`app-store-connect` 包总覆盖率 ≥ 70%
- [ ] `make gen` 一键重跑生成器，产物 byte-for-byte 复现
- [ ] `make update-spec` 一键同步 Apple 最新 spec
- [ ] Golden 清单快照完整，diff 可读
- [ ] README 更新反映新能力，保留原有示例 + gen Service 的使用示范
- [ ] `APP_STORE_CONNECT_API_PLAN.md` 勾选 Phase 6 全部子项

---

## 10. 未决问题

- Apple OpenAPI spec 的实际大小和字段命名规约需要在 Phase 6.1 开始时实际下载一份验证。如果实际情况和本设计的假设有出入（例如 spec 没有提供 examples、或者 resource 命名规则和预期不同），需要回到这份 spec 修订假设后再推进后续子阶段。
- 是否为常用 filter 参数生成类型安全的 options struct（例如 `AnalyticsReportsListOptions{Category: "APP_USAGE"}`）留到 Phase 6.3 跑过第一轮后再评估。初版先一律用 `*Query`。

---

## 11. 变更记录

| 日期 | 变更 |
|---|---|
| 2026-04-16 | 初稿，基于与用户的 5 轮 brainstorming（Q1-Q5）确定核心决策 |
