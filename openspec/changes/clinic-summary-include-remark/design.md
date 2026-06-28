## Context

`clinic_summary.go` 经 `DeviceHistory().ListHistory` 拉取近 7 天记录，输出 `[]clinicEventAgg` JSON 数组注入 system prompt。`entity.History.Remark` 已持久化但聚合循环未读取。基线 spec 要求「非全量 dump」——本变更在保留 `by_event` 聚合前提下，附加**有备注**记录的有限子集，与 `buildGrowthSuggestHistory` 的 `note` 语义对齐。

## Goals / Non-Goals

**Goals:**

- Hybrid JSON：`{ "by_event": [...], "records_with_remark": [...] }`
- `records_with_remark`：仅 `strings.TrimSpace(remark) != ""`；最多 30 条；按 `start_time` 降序
- 单条 `remark` 最长 200 字（超出截断，不追加省略号或追加 `…` 均可，实现统一即可）
- 更新 systemPrompt 引导 LLM 使用备注

**Non-Goals:**

- 7 天全量逐条 history（含空 remark）
- Flutter / WS 协议变更
- 隐私政策或 consent 文案（可后续单独 change）
- 新增 Redis 键或 DB 字段

## Decisions

### D1：Hybrid 结构，保留 `by_event`

用户明确要求「保留纯聚合」。`by_event` 字段与现有 `clinicEventAgg` 字段一致（`event_id/event_name/count/total_amount/amount_unit/total_duration_minutes/last_at`）。顶层由数组改为 object，breaking 仅影响 Redis 缓存 JSON 形态，无外部 API。

### D2：`records_with_remark` 筛选与排序

- 窗口：与聚合相同，7 天 cutoff（`start_time` 或 `end_time` 任一在窗口内即保留，与现有聚合一致）
- 筛选：`remark` trim 后非空
- 排序：`start_time` 降序（无 start 用 end）
- 截断：取前 30 条

### D3：单条记录字段

| 字段 | 规则 |
|------|------|
| `event_name` | 与聚合一致，优先 event 配置名 |
| `start_time` | `formatLocalDatetimeFromUnix`，复用 voice 包既有格式化 |
| `amount_value` | `event_number>0` 时输出，否则省略 |
| `duration_minutes` | start/end 有效时输出 |
| `remark` | trim + max 200 runes/bytes（实现用 `[]rune` 截断 200） |

### D4：Prompt 文案

`clinic_llm.go` 中摘要块标签由「近7天喂养事件聚合摘要（JSON）」改为「近7天喂养摘要（JSON，含 by_event 聚合与有备注记录）」；`config.voice-service.yaml` systemPrompt 同步说明备注块含义。

### D5：缓存兼容

旧缓存为纯数组 JSON；解析失败或检测到非 object 时视为 miss，触发重算。无需 migration 脚本。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 仅 30 条备注可能遗漏更早的关键备注 | 按时间降序保留最近 30 条；聚合仍覆盖全 7 天统计 |
| Token 略增 | cap 30 × 200 字；仍远小于全量 dump |
| 无备注时行为与现网一致 | `records_with_remark` 为空数组 `[]` |

## Migration Plan

1. 部署 voice-service 新版本
2. 下次 question 时 watermark 未变但 summary 结构变——可选：部署时不清缓存，首次命中旧数组格式时重算（D5）
3. 无回滚数据迁移；回滚代码后旧数组格式恢复
