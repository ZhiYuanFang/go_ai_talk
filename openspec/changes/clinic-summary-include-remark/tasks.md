## 1. 摘要结构（clinic_summary.go）

- [x] 1.1 定义 `clinicSummaryPayload`（`by_event` + `records_with_remark`）及 `clinicRemarkRecord` 结构体
- [x] 1.2 `buildClinicHistorySummary`：保留现有 by_event 聚合；扫描 7 天窗口内 remark 非空记录，构建 `records_with_remark`（start_time 降序，max 30，remark max 200 rune）
- [x] 1.3 单条记录字段：`event_name`、`start_time`、可选 `amount_value`/`duration_minutes`、`remark`；时间格式化复用 voice 包既有 helper

## 2. Prompt 与配置

- [x] 2.1 `clinic_llm.go`：更新 system 中摘要块标签文案
- [x] 2.2 `config.voice-service.yaml`：`aiClinic.systemPrompt` 说明 by_event 与有备注记录

## 3. 缓存兼容

- [x] 3.1 `ensureClinicSummary`：若 Redis 缓存 summary 为旧版纯数组 JSON，视为 miss 并重算

## 4. 验证

- [x] 4.1 `go build ./...` 通过
- [x] 4.2 `openspec validate clinic-summary-include-remark --strict` 通过
