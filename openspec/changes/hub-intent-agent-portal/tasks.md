## 1. 查史播报 Voice 代理

- [x] 1.1 `python_ai_client.go`：`ListHistorySpeakCache` / `BulkUpsertHistorySpeakCache` / `DedupeHistorySpeakCache` / `DeleteHistorySpeakCache` → `/v1/admin/history-speak-cache*`
- [x] 1.2 `api/v1`：`/voice/admin/api/history-speak-vectors*`（list/bulk/dedupe/delete）DTO
- [x] 1.3 `VoiceAdminHistorySpeakVectorsCtrl`（口令同 type-vectors）+ `register_voice_service` Bind

## 2. 门户与导航

- [x] 2.1 新增 `intent-agent-admin.html` 门户：四卡（事件向量 / 类型向量 / 查史播报向量 / 规则缺口）
- [x] 2.2 `admin-modules.js`：新增「意图智能体」`showInNav: true`；原三模块 `showInNav: false`；登记查史播报模块（`showInNav: false`）
- [x] 2.3 `admin_static_pages.go` + `gateway_app_auth_exempt.go`：门户 + 查史播报页
- [x] 2.4 三旧页 + 新查史页：回链「意图智能体」门户（最小文案）

## 3. 查史播报静态页

- [x] 3.1 新增 `history-speak-vector-admin.html`：list/bulk/dedupe/delete；payload 提示 `point_pick`/`open_speak`；集合名 `feeding_history_speak_cues`

## 4. 验收

- [x] 4.1 Hub 顶栏仅「意图智能体」；门户四链可进；旧三 URL 深链仍可用（代码已按此实现；部署后目视确认）
- [x] 4.2 Python history-speak-cache 已在兄弟仓注册；页内失败提示可读（部署联调确认 list/bulk）
- [x] 4.3 部署清单：voice-service + gateway-app；依赖 python-ai-talk admin 路由同发或先发
