## go_ai_talk

### 1. voice-service：voice_ai + clinic_ai 配额域

- [x] 1.1 DDL：在 `ai_voice_voice`（`VOICE_DB_LINK` / `database.default`）创建 `ai_quota_default`、`ai_quota_user_override` 表（含 `voice_ai_monthly_limit`、`clinic_ai_monthly_limit` 列；默认 5/30）
- [x] 1.2 DAO/entity/do：voice 域 ai_quota 代码生成或手写（`hack/config.dao.voice.yaml`）
- [x] 1.3 实现 `internal/services/voice/ai_quota.go`：全局默认、per-wxId override、Redis 用量（`ai:usage:voice_ai:*`、`ai:usage:clinic_ai:*`）、上海时区 YYYYMM
- [x] 1.4 Internal API：`POST /voice/internal/api/ai-quota/check|consume`（feature: `voice_ai`|`clinic_ai`）
- [x] 1.5 App API：`GET /voice/app/api/ai-quota` → `{ voiceAi, clinicAi }`
- [x] 1.6 Admin API：`GET/PUT /voice/admin/api/ai-quota/default|user`（`X-Admin-Password`）
- [x] 1.7 改造 voice chat/clinic WS：本地 check/consume，移除对 device ai-quota HTTP 客户端的依赖
- [x] 1.8 wxId-by-deviceNo：改调 device user 域 internal API（如 `InternalByID` / device-no 反查契约），删除 ai-quota 附带反查

### 2. ucg-service：polish 配额本地化

- [x] 2.1 DDL：在 `ai_voice_ucg`（`UCG_DB_LINK` / `database.default`）创建 `ai_quota_default`、`ai_quota_user_override`（仅 `polish_monthly_limit`）
- [x] 2.2 实现 `internal/services/ucg/ai_quota.go`：本地配置 + Redis `ai:usage:polish:*`
- [x] 2.3 改造 `POST /ucg/app/api/posts/polish`：进程内 check/consume，移除 device HTTP 转发
- [x] 2.4 App API：`GET /ucg/app/api/ai-quota` → `{ polish }`
- [x] 2.5 Admin API：本地化 `GET/PUT /ucg/admin/api/ai-quota/default|user`（**仅** polish 字段），删除转发 device 逻辑

### 3. device-service：完全退出 AI 配额

- [x] 3.1 删除 ai_quota DAO/entity、service、controller、api/v1 契约
- [x] 3.2 删除 `/device/internal/api/ai-quota/*`、`GET /device/app/api/ai-quota`、device admin ai-quota 接口
- [x] 3.3 维护窗口后 DROP device 库 `ai_quota_default`、`ai_quota_user_override` 表（或归档脚本）

### 4. gateway-app / gateway-service

- [x] 4.1 新增 `voice_route_proxy.go`：反代 `/voice/app/api/*`、`/voice/admin/api/*`（`VOICE_API_PROXY_URL`、`VOICE_API_ROUTE_MODE`）
- [x] 4.2 扩展 `InjectAdminDownstreamPassword`：`/voice/admin/api/` → `VOICE_ADMIN_PASSWORD`
- [x] 4.3 从 `device_route_proxy.go` **移除** `/device/app/api/ai-quota` 登记
- [x] 4.4 gateway-service 同步 voice HTTP 反代路径
- [x] 4.5 更新 `manifest/docker/.env.example`、`config.gateway-app-server.yaml` 注释与 `VOICE_API_PROXY_URL` 占位

### 5. Admin UI

- [x] 5.1 新建 `resource/public/voice-admin.html`（voiceAi + clinicAi 全局/用户配置）
- [x] 5.2 `admin-modules.js` 增加 `voice-admin` Hub 入口（`/device/admin/voice-admin.html`）
- [x] 5.3 `ucg-admin.html`：移除 voiceAi/clinicAi 字段与相关 JS

### 6. 配置与文档

- [x] 6.1 更新 `docs/runbooks/dao-sync-by-domain.md`：voice/ucg ai_quota 表归属
- [x] 6.2 更新 `docs/runbooks/release-deploy-and-run.md`：`VOICE_ADMIN_PASSWORD`、`VOICE_API_PROXY_URL`
- [x] 6.3 **App API usage 统计**：已与负责人确认 `GET /voice/app/api/ai-quota`、`GET /ucg/app/api/ai-quota` 是否计入统计；按结论更新 `maintenance_skip.go` 或 apiregistry
- [x] 6.4 确认无新增 `*_test.go`、无 background ticker

### 7. 验收（go_ai_talk）

- [ ] 7.1 手工验证：voice internal check/consume、App 读 API、voice-admin 配置
- [ ] 7.2 手工验证：ucg polish check/consume、App 读 API、ucg-admin 仅 polish
- [ ] 7.3 手工验证：gateway voice 反代、旧 `/device/app/api/ai-quota` 不可达、40301/40302 语义不变

---

## flutter_ai_talk

### 1. 模型与 Provider 拆分

- [x] 1.1 拆分 `aiQuotaStatusProvider` → `voiceAiQuotaProvider`（`GET /voice/app/api/ai-quota`）+ `polishAiQuotaProvider`（`GET /ucg/app/api/ai-quota`）
- [x] 1.2 更新 `ai_quota_models`：voice 响应 `{ voiceAi, clinicAi }`；polish 响应 `{ polish }`；移除聚合 device 模型
- [x] 1.3 移除对 `/device/app/api/ai-quota` 的全部调用与 env 路径引用

### 2. UI 更新

- [x] 2.1 `home_screen`：喂养 AI 额度展示改订阅 `voiceAiQuotaProvider`
- [x] 2.2 `pangbao_ai_screen`：胖宝额度改订阅 `voiceAiQuotaProvider`（clinicAi）
- [x] 2.3 UCG 润笔 UI：改订阅 `polishAiQuotaProvider`
- [x] 2.4 `AiQuotaRemainingHint`：按 feature 从对应 provider 取 used/limit

### 3. 错误处理（保持）

- [x] 3.1 HTTP/WS 40302 → 弹框「本月额度已用完」；40301 → 引导登录（voice + polish + clinic）

### 4. 验收（flutter_ai_talk）

- [ ] 4.1 端到端：首页喂养 AI、胖宝页、润笔页分别展示正确额度
- [ ] 4.2 额度用尽与未登录错误 UI 回归
- [ ] 4.3 确认无残留 `/device/app/api/ai-quota` 请求

---

## migration

### 维护窗口 Runbook（15–30 分钟）

- [ ] M.1 **窗口前**：备份 device `ai_quota_*` 表；导出 Redis `ai:usage:*` 快照；通知用户 AI 功能短暂不可用
- [ ] M.2 **停服或维护模式**：停止或隔离 device/voice/ucg 实例（或仅关闭 AI 相关路由）
- [ ] M.3 **MySQL 迁移**：
  - voice 库 INSERT：自 device 表迁移 `voice_ai_monthly_limit`、`clinic_ai_monthly_limit` 全局默认与 per-wxId override 行
  - ucg 库 INSERT：自 device 表迁移 `polish_monthly_limit` 全局默认与 per-wxId override 行
- [ ] M.4 **Redis 验证**：键名 `ai:usage:{feature}:{wxId}:{YYYYMM}` 不变；确认 voice/ucg 进程读写同一 Redis 集群；抽样对账 used 计数
- [ ] M.5 **部署顺序**：voice-service + ucg-service 新版本 → gateway（voice 反代、移除 device ai-quota）→ device-service（删除 ai-quota 代码路径）
- [ ] M.6 **冒烟**：internal check/consume、分域 App 读 API、admin 配置、润笔/喂养/胖宝各一条成功路径
- [ ] M.7 **App 发版**：Flutter 分域 provider 版本与后端同步上线
- [ ] M.8 **窗口后**：DROP device `ai_quota_*` 表；监控 40302 率与 Redis 键增长
- [ ] M.9 **回滚预案**：保留备份；回滚时恢复 device ai-quota 代码 + gateway `/device/app/api/ai-quota` 反代 + MySQL/Redis 自备份恢复
