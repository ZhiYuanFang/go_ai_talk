## 1. Voice 代理

- [x] 1.1 `python_ai_client.go`：`ListMissInbox(ctx, offset, limit, status, failKind)` → `GET /v1/admin/miss-inbox`；`DeleteMissInbox(ctx, id)` → `DELETE /v1/admin/miss-inbox/{id}`
- [x] 1.2 `api/v1`：`/voice/admin/api/miss-inbox` list DTO（offset/limit/status/fail_kind）+ delete DTO
- [x] 1.3 Admin Ctrl（口令同 rule-gaps）+ `register_voice_service` Bind；中文业务注释

## 2. Hub 静态页与登记

- [x] 2.1 新增 `miss-inbox-admin.html`：分页列表；status（默认 open）/fail_kind 筛选；删前 confirm；列含 utterance/reply/fail_kind/device_no/conversation_id/created_at；文案标明独立于 rule_gaps/时间飞轮
- [x] 2.2 非空 `device_no` 链到 `/device/admin/history/{deviceNo}`（URL 编码）；空显示「—」
- [x] 2.3 `admin-modules.js`（`showInNav: false`）+ `admin_static_pages.go` + `gateway_app_auth_exempt.go`

## 3. 门户

- [x] 3.1 `intent-agent-admin.html` 规则工单区加「覆盖缺口」卡（副文案可含答不出/缺失语境）；回链门户

## 4. 验收

- [x] 4.1 门户工单可见覆盖缺口；列表默认 open；删后列表消失（代码按此；部署后目视）
- [x] 4.2 有 device_no 可进 history 运维页；空 device_no 无链（部署后目视）
- [x] 4.3 部署清单：voice-service + gateway-app；依赖 python-ai-talk miss-inbox 同发或先发
