## Why

阿里云 Green 文本审核在违规时，`Data.Reason` 字段可能返回 **JSON 字符串**（含 `riskTips` 等人可读说明）。当前 `parseTextModeration` 将该字段 **原样** 写入 `AuditVerdict.Reason`，再经各 audit consumer 落库至 `reject_reason` 并展示给作者，导致 App/管理端出现整段 JSON 而非可读违规提示。

## What Changes

- 在 `green_client.go` 文本审核解析层，从 `Data.Reason` **提取 `riskTips`** 作为 `AuditVerdict.Reason`；解析失败或字段为空时回退默认文案「违规已下架」。
- 若 `Reason` 已是纯文本（非 JSON），**保持原样** 使用，兼容旧响应格式。
- **不修改** 图片/视频审核路径（图片仍用固定默认文案；视频 MVP 仍 pass）。
- **不修改** 管理端人工驳回、已有 DB 存量 `reject_reason` 数据。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-green-audit`：明确 Green 文本机审失败时写入 `reject_reason` 的文案 MUST 为用户可读的 `riskTips`（或等价纯文本），MUST NOT 将 API 返回的原始 JSON 字符串直接暴露给作者。

## Impact

- **代码**：`internal/services/ucg/green_client.go`（主改动）；下游 `audit_post` / `audit_comment` / `audit_profile_job` / `audit_chat` **无需改动**（均消费 `verdict.Reason`）。
- **API/DB**：字段名不变（`rejectReason` / `reject_reason`）；仅 **新产生** 的文本机审驳回文案内容变化。
- **服务**：仅 `ucg-service`。
- **客户端**：无需改协议；展示文案由后端修正。
