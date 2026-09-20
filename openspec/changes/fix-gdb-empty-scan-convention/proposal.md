## Why

GoFrame/MySQL 对单行 `.Scan(&struct)` 在空结果集时，部分驱动会返回 `sql.ErrNoRows`。仓库内多处把该错误当成系统失败（WARN、5xx、审核 MQ requeue），而实际业务语义是「无行 = 正常空」。已知触发点：`LoadLatestPendingProfileJob` 打出 `[ucg-profile] 读取待审 job 失败 … no rows`。本仓已在 `wx.go`、`force_store.go` 等处用 `.One()` + `IsEmpty()` 规避，但未形成全局约定，同类坑仍在 ucg/voice/cash/device/history/simuser 扩散。

## What Changes

- 将「可能无行」的单结构体查询统一改为 **`.One()` + `IsEmpty()`**（必要时再 `Struct`）；禁止把空集 `ErrNoRows` 当系统失败。
- 批量修复已审计的 **高危**（空集正常却传播 err）与 **中危**（`_ = Scan` 吞掉真 DB 错）调用点；**不改** `Scan(&[]T)` 列表语义，不改「插入后必有行」的回读。
- 在 **`openspec/project.md`** 新增强制约定，并在 **`AGENTS.md`** 增加短摘要指向，避免 AI/人工再次引入同类写法。
- 行为上属于 **修复误报/误失败**：对外 API 在「无可选行」场景恢复为成功空结果或明确业务错误，而非裸 `sql.ErrNoRows`；**无 BREAKING** 契约变更。

## Capabilities

### New Capabilities

- `gdb-empty-query-convention`：GoFrame 单行空结果查询约定——强制 `.One()` + `IsEmpty()`（或显式处理 `ErrNoRows`），禁止空集 Scan 当系统失败；含评审检查项与适用/排除边界。

### Modified Capabilities

- （无）本变更不修改既有领域 capability 的业务需求文本；v3.0.3 中已有个别场景（如原力空行插入、cash 订单空查）已约束不得抛裸 `ErrNoRows`，实现层对齐这些既有意图。

## Impact

- **代码**：`internal/services/{ucg,voice,cash,device,history,simuser}` 中约 25 处高危 + 约 20 处中危单行查询；canonical 样例：`LoadLatestPendingProfileJob`、审核 MQ loaders、Care Alert / 成长轨迹 latest、quota override、history 无记录等。
- **文档**：`openspec/project.md`、`AGENTS.md`。
- **API / DB / Redis / MQ**：无 schema 变更、无新接口、无新 Redis 键、无新背景循环任务；审核 consumer 对「job/帖/评已删」应 Ack 跳过而非因 `ErrNoRows` 无限 requeue。
- **gateway-app / usage 统计**：不新增 App HTTP 路由，无需询问 usage。
