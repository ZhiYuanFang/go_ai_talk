## 1. voice 代理 type-cache

- [x] 1.1 `python_ai_client.go`：实现 `ListTypeCache` / `BulkUpsertTypeCache` / `DedupeTypeCache` / `DeleteTypeCache` → `/v1/admin/type-cache*`
- [x] 1.2 `api/v1`：新增 `/voice/admin/api/type-vectors`（GET）、`/bulk`（POST）、`/dedupe`（POST）、`/{id}`（DELETE）请求/响应 DTO（风格对齐 intent-vectors）
- [x] 1.3 新增 `VoiceAdminTypeVectorsCtrl`（口令同 intent-vectors）；在 `register_voice_service` Bind

## 2. Hub 类型向量页与导航

- [x] 2.1 新增 `resource/public/type-vector-admin.html`：列表/分页、bulk（payload 提示 `target_type`/`op`/`history_mode`）、dedupe、删除；骨架对标 `intent-vector-admin.html`
- [x] 2.2 登记 `admin-modules.js`（标题「类型向量」）、`admin_static_pages.go`、`gateway_app_auth_exempt.go`
- [x] 2.3 页面 hint：数据在 Python `feeding_type_cues` / type-cache；非事件短说法库

## 3. 规则缺口页 TIME/QTY UX

- [x] 3.1 更新 `rule-gaps-admin.html`：默认不传 `dimension`；文案改为默认 TIME+QTY；可选下拉/提示支持显式 `time`/`quantity` 与 legacy 排障
- [x] 3.2 在页脚或 design/tasks 附注：存量 JSONL 可选运维清理（truncate 或只保留 time/quantity）；Hub 无清理按钮

## 4. 自检

- [x] 4.1 确认路径落在 `/voice/admin/api/*` 反代内；无 App 新路径、不改 `maintenance_skip`
- [x] 4.2 跑 `hack/check-service-import`；相关包可编译
- [x] 4.3 联调：type-cache 有数据时 Hub 可 list/bulk/删；规则缺口默认列表无 type/event/shape 噪音（依赖 Python 默认过滤）
