## Why

Python 兄弟仓（`python_ai_talk`）在多维本地规则未覆盖时已将缺口写入 JSONL 收件箱，并提供 `GET/PATCH /v1/admin/rule-gaps`；Go Hub 尚未代理与展示，运维在 Hub 找不到入口。需要按「意图向量缓存」同模式接上 Hub，能浏览并标已处理即可。

## What Changes

- voice-service 经既有 `PythonAIClient` 代理 Python `/v1/admin/rule-gaps`（列表 + 标 dismissed）。
- 新增 Hub 静态页与 `admin-modules.js` 导航（与「意图向量缓存」并列）。
- 页面支持按维度/状态筛选、分页查看整句/段/reason/LLM 摘要，并将条目标为已处理。
- **不**把缺口迁入 Go/MySQL；**不**热更新 Python 规则；**不**写事件向量库。

## Capabilities

### New Capabilities

- `hub-rule-gaps-admin`：Hub 规则缺口收件箱（代理 Python rule-gaps、列表/筛选/标已处理、导航入口）。

### Modified Capabilities

- （无。基线无已归档的 rule-gaps Hub 能力；Python 侧契约已由兄弟仓交付，本变更仅 Go Hub 接入。）

## Impact

- **voice-service**：`python_ai_client` 增 List/Patch；Admin 控制器 + `api/v1` DTO；鉴权同 voice Admin 口令。
- **gateway-app**：登记静态页与 auth exempt（若同 intent-vector 模式）；`/voice/admin/api/*` 反代已覆盖新 API。
- **Hub 静态资源**：`rule-gaps-admin.html`、`admin-modules.js`、`admin_static_pages.go`。
- **python_ai_talk**：只读调用既有 API，本仓不改 Python。
- **usage**：Admin 非 App 接口，不改 `maintenance_skip`。
