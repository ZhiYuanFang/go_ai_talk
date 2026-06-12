## 1. Green 文本驳回 Reason 解析

- [x] 1.1 在 `internal/services/ucg/green_client.go` 新增 `normalizeGreenRejectReason(raw string) string`：JSON 取 `riskTips`、纯文本兼容、失败回退空串
- [x] 1.2 在 `parseTextModeration` 中调用 normalize：labels 命中违规时，对 `Data.Reason` 规范化后再赋 `AuditVerdict.Reason`；空则 `rejectReasonDefault`
- [x] 1.3 补充中文注释说明：为何集中解析、JSON/纯文本/回退语义

## 2. 校验

- [x] 2.1 `go build ./...` 通过
- [x] 2.2 `openspec validate fix-green-risk-tips-reject-reason --strict` 通过
