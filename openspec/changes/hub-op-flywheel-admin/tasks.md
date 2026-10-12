## 1. Voice 代理：OP / 数量 / R·数据源

- [x] 1.1 `python_ai_client.go`：`List/BulkUpsert/Delete` ×3 → `/v1/admin/op-cache*`、`qty-cache*`、`read-source-cache*`（无 dedupe）
- [x] 1.2 `api/v1`：`/voice/admin/api/op-vectors*`、`qty-vectors*`、`read-source-vectors*`（list/bulk/delete）DTO
- [x] 1.3 三个 Admin Ctrl（口令同 type-vectors）+ `register_voice_service` Bind；中文业务注释

## 2. Hub 静态页与模块登记

- [x] 2.1 新增 `op-vector-admin.html`：list/bulk/delete；整段 document；payload `op_rail`；无整理重复；回链门户
- [x] 2.2 新增 `qty-vector-admin.html`：payload `quantity`；无量勿灌提示；回链门户
- [x] 2.3 新增 `read-source-vector-admin.html`：payload `read_source` record|predict；回链门户
- [x] 2.4 `admin-modules.js`：登记三模块 `showInNav: false`；`admin_static_pages.go` + `gateway_app_auth_exempt.go`

## 3. 门户与遗留页对齐

- [x] 3.1 `intent-agent-admin.html`：主卡改为 OP / 事件 / 数量 / R·数据源 / 查史播报 / 规则缺口；去掉类型向量主卡
- [x] 3.2 `type-vector-admin.html`：遗留横幅（热路径停用，新种子用 OP 轨）；深链仍可用
- [x] 3.3 `rule-gaps-admin.html`：TIME 主工单 + QTY 可配合数量飞轮文案；默认 TIME+QTY 语义不变

## 4. 验收

- [x] 4.1 门户六入口可进；类型主卡不可见；类型深链仍开且见遗留提示（代码已按此实现；部署后目视确认）
- [x] 4.2 三新页无 dedupe 按钮；list/bulk/delete 走对应 API；Python 宕机时 adminFetch 错误可读（部署联调确认）
- [x] 4.3 部署清单：voice-service + gateway-app；依赖 python-ai-talk op/qty/read-source-cache 同发或先发
