## Why

兄弟仓 `python_ai_talk`（变更 `time-dayword-flywheel-miss-inbox`）已在终局「答不出」前落独立收件箱，并挂好 `GET/DELETE /v1/admin/miss-inbox`。Go Hub 意图智能体门户尚无入口，运维看不到原话/失败回复，难以发现覆盖盲区。需按规则缺口同模式接上 Hub 查删，且 `device_no` 可跳转既有设备 history 运维页排查。

## What Changes

- voice-service 经 `PythonAIClient` 代理 Python `/v1/admin/miss-inbox`（list + 按 id delete；**无** PATCH/bulk）。
- 新增 Hub 静态页：分页列表、按 `status`/`fail_kind` 筛选、单条删除；非空 `device_no` 链到 `/device/admin/history/{deviceNo}`。
- 意图智能体门户「规则工单」区新增「覆盖缺口」卡片；登记模块 / 静态路由 / auth exempt。
- **不**与 rule_gaps / 时间飞轮混存；**不**改 Python 落库判定；**不**迁 JSONL 入 Go/MySQL。

## Capabilities

### New Capabilities

- `hub-miss-inbox-admin`：覆盖缺口（答不出 / 缺失语境）Admin 代理、Hub 查删页、门户入口与 device_no→history 深链。

### Modified Capabilities

- （无。既有 rule-gaps / 飞轮代理契约不改；门户仅新增工单卡。）

## Impact

- **voice-service**：`python_ai_client` List/Delete；Admin controller + `api/v1` DTO；鉴权同 voice Admin 口令。
- **gateway-app**：`admin-modules.js`、`admin_static_pages.go`、`gateway_app_auth_exempt.go`；门户 HTML；新静态页。
- **python_ai_talk**：只读调用既有 miss-inbox（须线上已发版）；本仓不改 Python。
- **usage**：Admin 非 App，不改 `maintenance_skip`。
- **部署**：`voice-service` + `gateway-app`。
