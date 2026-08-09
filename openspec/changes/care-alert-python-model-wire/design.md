## Context

兄弟仓 `python_ai_talk` 的 `POST /v1/care-alert/analyze` 请求体字段为：

- `device_no` / `day` / `age_months` / `history_summary` / `kg_context`
- **`model`**（可选）：`None` | `"deepseek"|"zhipu"` | `{provider, name, max_in_flight}`

编排仅调用 `resolve_model_config(request.model)`；**无 `model_cfg` 字段**。

本仓当前 `CareAlertAnalyzeRequest` 含：

- `model`：`string`（omitempty，调用方未赋值）
- `model_cfg`：`*PythonModelCfg`（有选模结果时赋值）

权益路径 `ResolveLaneModel` 工作正常；断点在线格式。clinic/tip/intent 已统一使用 `model: *PythonModelCfg`。

## Goals / Non-Goals

**Goals:**

- premium / 已配置 free 时，Go 请求体携带 `model` 对象，Python 日志可见首选 `provider`/`name`。
- 非 premium 且 free 空时 omit `model`，Python 走保底序。
- 与 clinic/tip/intent 的 JSON 形状一致；删除对 `model_cfg` 的依赖。
- CONTRACT / 注释与真实契约一致。

**Non-Goals:**

- 不改 Python 仓（已为真相源）。
- 不改 VIP/额度判定、`careAlert` lane 配置、计次、日缓存。
- 不恢复旧的「仅传 provider 字符串」为 Go 唯一路径（对象更完整；Python 仍兼容字符串）。

## Decisions

### D1：权威字段为 `model` 对象

- **选择**：`CareAlertAnalyzeRequest.Model *PythonModelCfg \`json:"model,omitempty"\``；删除 `ModelCfg`/`model_cfg` 与字符串 `Model`。
- **理由**：与 Python `ModelConfig` / clinic 一致；避免双字段漂移。
- **备选**：继续填字符串 `model` + 另发 `model_cfg` → 拒绝（Python 不读后者；双写易再漂）。

### D2：调用方只映射 `ResolveLaneModel` 的指针

- **选择**：`CareAlertAnalyze(..., Model: modelCfg)`；`modelCfg==nil` 时 omit。
- **理由**：选模出口已表达 premium/free/omit；调用方不再拼 provider 简写。

### D3：文档同步

- **选择**：更新 `llm-care-alert-daily/CONTRACT.md` Go→Python 小节；结构体/服务中文注释写明与 Python 对齐。
- **备选**：仅改代码 → 拒绝（CONTRACT 仍写 `model_cfg` 会误导后续）。

## Risks / Trade-offs

- [旧 Python 若仍读 `model_cfg`] → 当前兄弟仓已不读；若环境未部署新 Python 则本就是旧简写契约，对象 `model` 亦被 `_normalize_model_field` 接受。风险低。
- [日缓存已用保底结果] → 修线后需清当日 Redis 日缓存或等次日才能看到首选模型效果；文档一句即可。

## Migration Plan

1. 部署 voice-service（仅本变更相关文件）。
2. 联调：premium 账号 miss 日缓存 → Python 日志非 `(fallback-only)`。
3. 回滚：回退 voice-service 版本；无 DDL。

## Open Questions

- 无阻塞问题。若运维需立刻验证，可手工删 `care-alert` 日缓存键后重拉 daily。
