## Context

GoFrame `Model.Scan(&T)` 在零行匹配时，MySQL 驱动常返回 `sql.ErrNoRows`。本仓大量「可选行」查询（按 wx_id / status / Limit(1) 取最新）把任意 `err` 当失败，导致：

- 作者主页/发帖回包：无 pending 资料 job 却打 WARN（`LoadLatestPendingProfileJob`）
- 审核 MQ：job/帖已删时本应 Ack，却因 Scan err requeue
- 首次无行业务（Care Alert / 成长轨迹 / quota override）误失败

仓库已有正确样例：`device/wx.go`（注释明确禁 Scan 空集）、`ucg/force_store.go`（`.One()` + `IsEmpty()`）。`openspec/project.md` / `AGENTS.md` 尚无对应强制约定。

本变更按 explore 方案 2：高危全改 + 中危可选查找改为 One + 全局规范落地。

## Goals / Non-Goals

**Goals:**

- 统一「可能无行」单行查询为 `.One()` → 真错误返回 → `IsEmpty()` 表示无行 → 再 `Struct`。
- 修复已盘点的高危与中危调用点，消除空集噪音 WARN / 误 5xx / 误 requeue。
- 将约定写入 `openspec/project.md` 与 `AGENTS.md`，评审可对照。

**Non-Goals:**

- 不引入新 DAO 抽象层 / 不封装通用 helper（除非实现时发现重复样板过多再评估）。
- 不改 `Scan(&[]T)` 列表查询语义。
- 不强制改「刚 Insert 后回读」或「Ensure*Row 后必有行」路径。
- 不做全自动 lint 门禁（可选后续 `hack/check-*.sh`）；本变更以约定 + 手工盘点清单验收。
- 不新增 Redis、不新增背景循环、不新增 App HTTP 接口。

## Decisions

### D1：首选 `.One()` + `IsEmpty()`，而非到处 `errors.Is(ErrNoRows)`

- **选择**：与现有 `wx.go` / `force_store.go` 一致；`One` 的空结果通常不抛 `ErrNoRows`，语义清晰。
- **备选**：保留 `Scan` 并显式 `errors.Is(err, sql.ErrNoRows)` —— 可接受为遗留过渡，但本变更新改代码优先 One。
- **排除**：继续 `_ = Scan(...)` 吞错 —— 空集碰巧 OK，真 DB 故障也被吞。

### D2：空集语义分层

| 场景 | 改造后语义 |
|------|------------|
| 可选最新行（pending job、latest、override） | `(ok=false)` / 零值 / 默认配置，**不** WARN |
| 按 ID 加载且「可能已删」（审核 MQ） | 视为 missing → **Ack/skip**，不 requeue |
| 业务上「必须存在」 | 返回明确业务错误（NotFound 等），**禁止**裸 `ErrNoRows` 出站 |
| 列表 / 分页 `Scan(&[]T)` | 不动 |

### D3：规范双写位置

- **权威**：`openspec/project.md` 新增「GoFrame 单行空结果查询约定（强制）」节（正反例、MUST/禁止、评审项）。
- **执行摘要**：`AGENTS.md` 短列表 + 指向 `project.md`（对齐 Redis 约定写法）。

### D4：修复范围按域分批，同一 PR / 同一 change

- 分域 tasks（ucg → voice → cash → device/history → simuser → docs），便于评审与回滚定位；不拆多个 OpenSpec change。

### D5：不新增 usage / gateway / Redis 询问项

- 无新 App 路由、无新 Redis 读缓存；tasks 中注明 N/A。

## Risks / Trade-offs

- **[漏改]** → 以 explore 盘点清单为 tasks 勾选表；apply 结束前再 `rg '\.Scan\(&' internal/services` 人工复核单结构体调用。
- **[行为变化被误认为回归]** → Care Alert / quota 等「首次无行」从失败变成功，是修复；PR 说明中写明。
- **[审核 loaders 改 Ack]** → 若 ID 永远无效（毒消息）应 Ack；若短暂读副本延迟极罕见，本仓单主库无该问题。
- **[改动面大]** → 仅错误语义，无 schema；可按域提交但保持同一 change。

## Migration Plan

1. 合并后滚动发布各服务即可，无数据迁移。
2. 回滚：恢复旧 Scan 逻辑即可；无 DB/Redis 副作用。
3. 验证：打 `GET profile/me` / 发帖回包，确认不再出现 `读取待审 job 失败 … no rows`；可选构造已删 jobId 的 MQ 消息确认 Ack。

## Open Questions

- （无阻塞项）自动化 `hack/check-gdb-empty-scan` 是否在本 change 做：默认 **不做**；若实现中样板极重复可再议。
