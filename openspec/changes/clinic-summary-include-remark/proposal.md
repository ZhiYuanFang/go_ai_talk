## Why

胖宝诊疗（clinic）当前注入 LLM 的近 7 天喂养摘要仅含按事件类型的聚合统计（count/amount/duration），**丢弃了用户为每条记录填写的备注**。备注常含侧别、宝宝状态、异常描述等关键上下文，缺失会导致 AI 回答偏离实际情况。history 表与 App 已支持 `remark` 字段，语音球成长建议路径亦已逐条携带 `note`，clinic 应补齐同等语义。

## What Changes

- 扩展 `buildClinicHistorySummary` 输出：**保留**现有 `by_event` 聚合；**新增** `records_with_remark` 数组，仅收录近 7 天内 **`remark` 非空** 的记录，最多 **30** 条（按开始时间降序截断）。
- 每条 `records_with_remark` 含：`event_name`、`start_time`（本地 `YYYY-MM-DD HH:mm:ss`）、可选 `amount_value`/`duration_minutes`、`remark`（截断至 200 字）。
- 更新 `aiClinic.systemPrompt`：说明结合聚合统计与用户备注记录回答，勿编造未出现的数据。
- **不变**：WS 协议、Flutter 客户端、Redis 缓存键/watermark 懒刷新机制、history HTTP 契约；**非** 7 天全量逐条 dump（无备注记录仍仅出现在聚合中）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `pangbao-ai-clinic`：修订「近 7 天喂养事件聚合摘要」需求，允许在聚合之外附加有备注的记录子集（≤30 条）。

## Impact

- **代码**：`internal/services/voice/clinic_summary.go`、`clinic_llm.go`（prompt 标签文案可选微调）、`manifest/config/config.voice-service.yaml`（`aiClinic.systemPrompt`）。
- **缓存**：`voice:clinic:summary:*` JSON 结构由数组变为对象；watermark 逻辑不变，旧缓存自然过期重算。
- **对外 API/WS**：无协议变更。
- **Token**：有备注记录 capped 30 条，单条 remark ≤200 字，可控于现有 clinic 上下文预算。
