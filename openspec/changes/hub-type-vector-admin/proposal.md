## Why

兄弟仓 `python_ai_talk`（变更 `type-keyword-flywheel`）已交付类型关键字向量库与 `/v1/admin/type-cache*`，并把规则缺口收窄为仅 TIME/QTY；Go Hub 仍缺「类型向量」整理入口，且规则缺口页文案仍暗示「空=全部维度」，与 Python 默认列表语义不一致。需对标「意图向量缓存」补齐 Hub，并收口缺口页展示语义。

## What Changes

- voice-service 经既有 `PythonAIClient` 代理 Python `/v1/admin/type-cache`（list / bulk / dedupe / delete）。
- 新增 Hub 静态页「类型向量」与 `admin-modules.js` 导航（对标 `intent-vector-admin`）。
- 规则缺口页：默认展示语义对齐 TIME/QTY（文案与筛选控件）；不把 type/event/shape 当默认工单箱。
- **可选运维**：存量 `gaps.jsonl` 中 legacy 维度行的清理——本变更仅提供 runbook/说明，**不**新增 Hub 热清理按钮，**不**在 Go 进程读写 Python JSONL。
- **不**改 Python 飞轮/写路径；**不**把类型向量或缺口迁入 Go/MySQL。

## Capabilities

### New Capabilities

- `hub-type-vector-admin`：Hub 类型向量整理（代理 type-cache、list/bulk/dedupe/delete、静态页与导航）。
- `hub-rule-gaps-time-qty-ux`：规则缺口页默认 TIME/QTY 展示语义与筛选说明（对接已存在的 rule-gaps 代理）。

### Modified Capabilities

- （无。`hub-rule-gaps-admin` 尚未归档进基线 specs；本变更以增量 UX 能力描述，不改已交付的 list/dismiss 契约。）

## Impact

- **voice-service**：`python_ai_client` 增 type-cache 四方法；Admin 控制器 + `api/v1` DTO；鉴权同 voice Admin 口令。
- **gateway-app**：登记静态页与 auth exempt；`/voice/admin/api/*` 反代已覆盖新 API。
- **Hub 静态资源**：新建 `type-vector-admin.html`；调整 `rule-gaps-admin.html`；`admin-modules.js`、`admin_static_pages.go`。
- **python_ai_talk**：只读调用既有 type-cache / rule-gaps API；JSONL 可选清理由运维在 Python 数据卷执行。
- **usage**：Admin 非 App 接口，不改 `maintenance_skip`。
