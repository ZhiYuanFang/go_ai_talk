## Why

兄弟仓 `python_ai_talk` 已交付并挂好 Admin 的四套整段飞轮——时间相对配方（`time-cache`）、备注（`remark-cache`）、R·predict 两维（`predict-cache`）、U/D 定位（`locate-cache`）。既有 Go Hub 变更 `hub-op-flywheel-admin` 仅覆盖 OP/数量/R·source，且门户仍写「TIME 无飞轮」。运维无法按热路径灌种与巡检新增维度，文案也与上游事实冲突。

## What Changes

- voice-service 经 `PythonAIClient` 代理四套 Python Admin：`/v1/admin/time-cache*`、`remark-cache*`、`predict-cache*`、`locate-cache*`（list / bulk / delete；**无** dedupe，与上游一致）。
- 新增四张 Hub 静态整理页（镜像 OP/数量页骨架）并登记模块 / 静态路由 / auth exempt。
- 意图智能体门户飞轮区补卡：时间、备注、R·predict、U/D 定位；去掉「TIME 无飞轮」表述。
- 规则缺口页文案：TIME 仍为主工单，但可配合时间飞轮；不再暗示「时间维禁止飞轮」。
- **不**改 Python 飞轮热路径或配方校验；**不**迁 Chroma 入 Go/MySQL；**不**为新四层补 dedupe。

## Capabilities

### New Capabilities

- `hub-time-vector-admin`：时间相对配方飞轮 Admin 代理与 Hub 整理页（`feeding_time_cues`）。
- `hub-remark-vector-admin`：备注整段飞轮 Admin 代理与 Hub 整理页（`feeding_remark_cues`）。
- `hub-predict-vector-admin`：R·predict 两维飞轮 Admin 代理与 Hub 整理页（`feeding_predict_cues`）。
- `hub-locate-vector-admin`：U/D 定位策略飞轮 Admin 代理与 Hub 整理页（`feeding_locate_cues`）。
- `hub-flywheel-portal-time-align`：门户与规则缺口页对齐「已有 TIME 飞轮」文案，并暴露四新入口。

### Modified Capabilities

- （无。既有 OP/qty/read-source/history-speak/rule-gaps 代理契约不改；门户与缺口页展示语义以本变更新能力描述。）

## Impact

- **voice-service**：`python_ai_client` 增四组 list/bulk/delete；Admin controller + `api/v1` DTO；鉴权同 voice Admin 口令。
- **gateway-app**：`admin-modules.js`、`admin_static_pages.go`、`gateway_app_auth_exempt.go`；门户与规则缺口 HTML；四张新静态页。
- **python_ai_talk**：只读调用既有 admin（须线上已发版）；本仓不改 Python。
- **usage**：Admin 非 App，不改 `maintenance_skip`。
- **部署**：`voice-service` + `gateway-app`；依赖 python-ai-talk 四套 `*-cache` 已上线。
