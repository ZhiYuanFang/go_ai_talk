## ADDED Requirements

### Requirement: Green text rejection reason SHALL expose human-readable riskTips

当 Green **文本**机审判定违规（`Labels` 非空且非 `nonLabel`）时，系统 MUST 将 `AuditVerdict.Reason` 设为用户可读文案：**若 API `Data.Reason` 为 JSON 字符串，MUST 解析并取 `riskTips` 字段**；若 `Reason` 为纯文本，MUST 使用该纯文本；若解析后仍为空，MUST 使用默认文案「违规已下架」（或代码中等价常量）。系统 MUST NOT 将 `Data.Reason` 的原始 JSON 字符串直接写入 `reject_reason` 或经 App/WS 暴露给作者。

该规则 MUST 适用于所有经 `ModerateText` / `parseTextModeration` 的实体（帖子、评论、资料 job、私信）。图片/视频机审路径不在本 Requirement 范围内。

#### Scenario: Reason 为 JSON 时展示 riskTips

- **WHEN** Green 文本审核返回 `Labels` 命中且 `Data.Reason` 为 `{"riskTips":"命中违禁内容","riskLevel":"high"}`
- **THEN** 机审 consumer 写入的 `reject_reason`（及 WS `audit_failed` 的 reason）MUST 为 `命中违禁内容`，MUST NOT 为完整 JSON 字符串

#### Scenario: Reason 为纯文本时原样使用

- **WHEN** Green 返回 `Data.Reason` 为纯文本 `内容不合规`
- **THEN** `AuditVerdict.Reason` MUST 为 `内容不合规`

#### Scenario: JSON 无 riskTips 时回退默认文案

- **WHEN** Green 返回 `Data.Reason` 为 `{"riskLevel":"high"}` 且无有效 `riskTips`
- **THEN** `AuditVerdict.Reason` MUST 为「违规已下架」（或等价默认文案）

#### Scenario: 帖子作者可见修正后的驳回原因

- **WHEN** 用户帖子文本机审失败且 CAS 驳回成功
- **THEN** 作者「我的动态」接口返回的 `rejectReason` MUST 为上述可读文案，MUST NOT 含 JSON 花括号包裹的原始 Reason
