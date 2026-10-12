## Why

兄弟仓 `python_ai_talk`（变更 `intent-crud-other-reframe`）已将意图热路径切到 **C/R/U/D OP 轨 + 分层整段飞轮**，并挂好 `op-cache` / `qty-cache` / `read-source-cache` Admin；短关键词 TYPE 飞轮热路径已停用。Go Hub「意图智能体」门户仍暴露「类型向量」并缺少 OP/数量/R·数据源整理入口，运维无法按热路径维度灌种与巡检。需对齐 Python 已落地的飞轮面，并保留规则缺口（TIME/QTY）工单入口。

## What Changes

- voice-service 经 `PythonAIClient` 代理三套 Python Admin：`/v1/admin/op-cache*`、`/v1/admin/qty-cache*`、`/v1/admin/read-source-cache*`（list / bulk / delete；**无** dedupe，与上游一致）。
- 新增三张 Hub 静态整理页（镜像既有向量页）：OP 轨、数量、R·数据源；登记模块 / 静态路由 / auth exempt。
- 重排「意图智能体」门户卡：飞轮区为 OP / 事件 / 数量 / R·数据源 / 查史播报；工单区保留规则缺口；**门户不再展示「类型向量」主卡**（深链 URL 保留，页头标明遗留/迁移）。
- 规则缺口页文案微调：TIME 为主工单；QTY 可配合数量飞轮，不暗示「数量只能改规则」。
- **不**实现 R·predict / U·D 定位飞轮 Hub（Python 尚无对应 admin）；**不**改 Python 飞轮热路径；**不**迁 Chroma/JSONL 入 Go/MySQL。

## Capabilities

### New Capabilities

- `hub-op-vector-admin`：OP 轨整段飞轮 Admin 代理与 Hub 整理页（`op_rail`）。
- `hub-qty-vector-admin`：数量整段飞轮 Admin 代理与 Hub 整理页（`quantity`）。
- `hub-read-source-vector-admin`：R·数据源飞轮 Admin 代理与 Hub 整理页（`record|predict`）。
- `hub-flywheel-portal-align`：意图智能体门户按热路径飞轮重排；类型向量主入口降权；规则缺口文案对齐。

### Modified Capabilities

- （无。既有 intent/history-speak/rule-gaps 代理契约不改；门户与类型页展示语义以本变更新能力描述。）

## Impact

- **voice-service**：`python_ai_client` 增三组 list/bulk/delete；Admin controller + `api/v1` DTO；鉴权同 voice Admin 口令。
- **gateway-app**：`admin-modules.js`、`admin_static_pages.go`、`gateway_app_auth_exempt.go`；门户与类型页 HTML；三张新静态页。
- **python_ai_talk**：只读调用既有 op/qty/read-source-cache API（任务 5.3 兄弟仓侧）；本仓不改 Python。
- **usage**：Admin 非 App，不改 `maintenance_skip`。
