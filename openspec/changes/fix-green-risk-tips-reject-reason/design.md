## Context

- Green 文本审核 API（`TextModeration`）在命中标签时返回 `Data.Labels` 与 `Data.Reason`。
- 当前 `parseTextModeration`（`internal/services/ucg/green_client.go`）将 `Reason` trim 后直接赋给 `AuditVerdict.Reason`。
- 线上观测：`Reason` 为 JSON 字符串，例如 `{"riskTips":"…","riskLevel":"high",…}`，导致帖/评/资料/私信四类 consumer 写入 DB 与 WS 的 `reject_reason` 为整段 JSON。
- 下游 audit 路径（`audit_post.go`、`audit_comment.go`、`audit_profile_job.go`、`audit_chat.go`）均只读 `verdict.Reason`，**无需逐处修改**。

## Goals / Non-Goals

**Goals:**

- 文本机审失败时，`AuditVerdict.Reason` MUST 为可读中文提示（优先 `riskTips`）。
- 兼容 `Reason` 为纯文本的旧格式。
- 空/无效时回退 `rejectReasonDefault`（「违规已下架」）。

**Non-Goals:**

- 图片/视频审核 Reason 解析（图片仍固定默认文案）。
- 回填历史 DB 中已存的 JSON 字符串 `reject_reason`。
- 管理端人工驳回逻辑变更。
- 新增测试文件（仓库当前阶段约定）。

## Decisions

### 1. 解析入口集中在 `green_client.go`

**选择**：新增 `normalizeGreenRejectReason(raw string) string`，在 `parseTextModeration` 调用。

**理由**：单一 choke point，四类实体自动受益；与现有 `rejectReasonDefault` 常量同包。

**算法**：

1. `raw = strings.TrimSpace(raw)`；空 → 返回 `""`（由调用方决定是否用 default）。
2. 若 `raw` 以 `{` 开头，尝试 `json.Unmarshal` 到 `map[string]interface{}` 或小型 struct `{ RiskTips string \`json:"riskTips"\` }`。
3. 取 `riskTips`（或 `RiskTips`）trim 后非空 → 返回。
4. 否则若 `raw` 非 JSON 或解析失败 → 若 `raw` 不以 `{` 开头，返回 `raw`（纯文本兼容）；若以 `{` 开头但无有效 `riskTips` → 返回 `""`。

**备选**：在各 `audit_*.go` 落库前解析 — 拒绝，重复且易漏。

### 2. 不尝试从 `Labels` 反推文案

Green 的 `Labels` 为机器标签（如 `spam`），非用户可读；仅使用 `Reason`/`riskTips`。

### 3. 图片路径保持不变

`parseImageModeration` 继续返回 `rejectReasonDefault`；若后续 API 在 `Result` 中提供类似 JSON，可另开变更。

## Risks / Trade-offs

- **[Risk] JSON 结构字段名变化**（如 `risk_tips`）→ 先只支持文档确认的 `riskTips`；解析失败回退默认文案，不暴露原始 JSON。
- **[Risk] 存量 reject_reason 仍为 JSON** → 仅影响新驳回；可选后续脚本清洗，本变更不做。
- **[Trade-off] 纯 JSON 且无 riskTips 时用户只见默认文案** → 优于展示整段 JSON；日志可保留原始 `Reason` 供运维（若现有 Green 调用已有 error log 则复用，不强制新增）。

## Migration Plan

- 部署新版 `ucg-service` 即可；无 DDL、无 MQ 拓扑变更。
- 回滚：还原 `green_client.go` 解析逻辑。

## Open Questions

- 无。字段名 `riskTips` 已由线上响应确认。
