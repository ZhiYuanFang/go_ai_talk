## Why

`vip-quota-joint-entitlement` 落地后，care-alert 生成在权益判定为 premium（有 VIP 或 `care_alert` 额度）时仍未把首选模型交给 Python：Go 把配置写在 JSON 字段 `model_cfg`，而兄弟仓 `python_ai_talk` 的 `CareAlertAnalyzeRequest` **只解析 `model`**（字符串或 `{provider,name,max_in_flight}` 对象），忽略多余字段。结果 Python 日志出现 `(fallback-only)`，与 Go 侧「有额度」不一致。

## What Changes

- 将 Go `CareAlertAnalyzeRequest` 与 clinic/tip/intent 对齐：首选模型走 JSON 字段 **`model`**（`*PythonModelCfg`，`omitempty`）。
- **BREAKING（内部契约）**：不再向 Python 发送 `model_cfg`；不再使用已废弃的 `model` 字符串简写字段（`deepseek|zhipu`）作为权威路径（Python 仍可接受字符串，但 Go 统一发对象）。
- 生成路径：`ResolveLaneModel` 得到的 `modelCfg` 非 nil 时填入 `Model`；nil 时整段 omit，由 Python 走免费保底序（与 `llm-fallback-chain` 一致）。
- 更新本仓 `llm-care-alert-daily` CONTRACT 表述，去掉「`model` 简写 + `model_cfg`」误导。

## Capabilities

### New Capabilities

- `care-alert-python-wire`：Go → Python `/v1/care-alert/analyze` 模型字段线格式（`model` 对象或 omit；禁止依赖 `model_cfg`）。

### Modified Capabilities

- （无基线 `openspec/specs/` 独立 capability 文件需 delta；行为增量落在本变更 `care-alert-python-wire`。相关表述同步修正 `llm-care-alert-daily/CONTRACT.md`。）

## Impact

- **进程**：`voice-service`（`python_ai_client.go`、`care_alert_service.go`）。
- **跨仓**：与已部署的 `python_ai_talk` care-alert 分析接口对齐；**不要求**改 Python。
- **不改**：额度/`care_alert` feature、lane 选模逻辑、日缓存键、gateway 路由、App 对外 HTTP。
- **文档**：`openspec/changes/llm-care-alert-daily/CONTRACT.md`；可选一句 runbook。
