## 1. Voice 代理：时间 / 备注 / 预测 / 定位

- [x] 1.1 `python_ai_client.go`：`List/BulkUpsert/Delete` ×4 → `/v1/admin/time-cache*`、`remark-cache*`、`predict-cache*`、`locate-cache*`（复用 cue 通用方法；无 dedupe）
- [x] 1.2 `api/v1`：`/voice/admin/api/time-vectors*`、`remark-vectors*`、`predict-vectors*`、`locate-vectors*`（list/bulk/delete）DTO
- [x] 1.3 四个 Admin Ctrl（口令同 op-vectors）+ `register_voice_service` Bind；中文业务注释

## 2. Hub 静态页与模块登记

- [x] 2.1 新增 `time-vector-admin.html`：list/bulk/delete；整段 document；相对配方提示（禁 Unix 主缓存）；无整理重复；回链门户
- [x] 2.2 新增 `remark-vector-admin.html`：payload `remark`（可空）；回链门户
- [x] 2.3 新增 `predict-vector-admin.html`：`predict_channel` + `time_answer` 枚举提示；回链门户
- [x] 2.4 新增 `locate-vector-admin.html`：`strategy` 枚举；禁 historyId；回链门户
- [x] 2.5 `admin-modules.js`：登记四模块 `showInNav: false`；`admin_static_pages.go` + `gateway_app_auth_exempt.go`

## 3. 门户与规则缺口文案

- [x] 3.1 `intent-agent-admin.html`：飞轮区加时间 / 备注 / R·predict / U/D 定位；去掉「TIME 无飞轮」表述
- [x] 3.2 `rule-gaps-admin.html`：TIME 主工单 + 可配合时间飞轮；去掉「无时间飞轮」

## 4. 验收

- [x] 4.1 门户可见四新入口；TIME 文案已对齐；规则缺口无「无时间飞轮」（代码按此实现；部署后目视确认）
- [x] 4.2 四新页无 dedupe；list/bulk/delete 走对应 API；Python 宕机时 adminFetch 错误可读（部署联调）
- [x] 4.3 部署清单：voice-service + gateway-app；依赖 python-ai-talk time/remark/predict/locate-cache 同发或先发
