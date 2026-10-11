## Why

Hub 顶栏把「意图向量缓存 / 类型向量 / 规则缺口」摊成三个平级入口，运维找工具乱、命名也易与「意图智能体」混淆。同时 Python 已落地查史播报飞轮集合（`feeding_history_speak_cues`），Hub 尚无整理入口。需要门户收口 + 补齐查史播报向量页。

## What Changes

- Hub 顶栏仅保留一个入口：**意图智能体**，指向门户页；原三个模块 `showInNav: false`，仍可深链打开。
- 新增门户页：卡片/链接聚合——事件向量、类型向量、查史播报向量、规则缺口（标题纠正：事件短说法不再叫「意图向量缓存」作门户主名）。
- 新增查史播报向量：voice 代理 Python `/v1/admin/history-speak-cache*`（list/bulk/dedupe/delete）+ Hub 静态页（镜像类型向量页；payload `point_pick` / `open_speak`）。
- 子页面包屑/回链指向门户；旧 URL 保留不删。
- **依赖兄弟仓**：Python 须暴露 `history-speak-cache` Admin 路由（store 已有；若未挂路由须同发版）。本仓按既定契约对接，不直连 Chroma。

## Capabilities

### New Capabilities

- `hub-intent-agent-portal`：意图智能体门户与导航收口。
- `hub-history-speak-vector-admin`：查史播报向量 Admin 代理与 Hub 整理页。

### Modified Capabilities

- （无。既有 intent/type/rule-gaps 代理行为不改，仅导航与门户展示收口。）

## Impact

- **gateway-app**：`admin-modules.js`、`admin_static_pages.go`、`gateway_app_auth_exempt.go`；新门户 + 查史播报 HTML。
- **voice-service**：`PythonAIClient` + Admin controller + `api/v1`（history-speak-vectors）；`/voice/admin/api/*` 反代已覆盖。
- **静态页**：`intent-agent-admin.html`（门户）、`history-speak-vector-admin.html`；微调既有三页回链文案（可选最小改）。
- **python_ai_talk**：须有 `/v1/admin/history-speak-cache*`（孪生 `history-speak-flywheel`）；本仓不改 Python 代码。
- **usage**：Admin 非 App，不改 `maintenance_skip`。
