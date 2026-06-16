## go_ai_talk

### 1. device-service：clinic_ai 额度

- [x] 1.1 DDL/配置：ai_quota 全局默认与 per-wxId override 增加 `clinicAiMonthlyLimit`，默认值 **30**（进程 `device-service`，`database.default`）
- [x] 1.2 Internal API：`/device/internal/api/ai-quota/check|consume` 的 `feature` 枚举增加 `clinic_ai`；Redis 键 `ai:usage:clinic_ai:{wxId}:{YYYYMM}`
- [x] 1.3 App API：`GET /device/app/api/ai-quota` 响应增加 `clinicAi: { used, limit }`
- [x] 1.4 Admin：device Admin UI / ucg-admin「AI 配置」增加第三字段「胖宝 AI 月度次数」，PUT default/user 支持 `clinicAiMonthlyLimit`

### 2. voice-service：配置与 clinic 模块

- [x] 2.1 在 `manifest/config/config.voice-service.yaml` 新增 `aiClinic:` 块（`model: deepseek-v4-pro`、`llmTimeoutSeconds: 120`、thinking/reasoning 配置、rate limit 参数）；**不**写入 `voice-chat.shared.yaml`
- [x] 2.2 实现 `buildClinicHistorySummary`：经 `DeviceHistory().ListHistory` 拉取 7 天数据，按 event 聚合 count/amount/duration
- [x] 2.3 实现摘要懒刷新：Redis `voice:clinic:summary:{wxId}:{deviceNo}`，对比 history watermark 决定是否重算
- [x] 2.4 实现 session：Redis **`voice:clinic:session:{wxId}`**（wxId 主键，非 deviceNo），首问创建、12h 固定 TTL（非 sliding），存多轮 Q&A 与 auth 锁定 `deviceNo`
- [x] 2.5 实现 rate limit：Redis 键 **`voice:clinic:rate:{wxId}`**，超限返回 WS 42901
- [x] 2.6 实现 DeepSeek 流式调用：解析 thinking/content 双流，映射 `thinking_delta` / `answer_delta` / `answer_done`
- [x] 2.7 实现 `/voice/clinic/ws` handler：首帧 `auth` 解析 JWT 得 `wxId>0` 与 `deviceNo`（须与 JWT 一致）；`auth_ok` 后处理 `question`；LLM 前 `clinic_ai` check、成功后 consume；**不**注册 `VoiceWSManager`；**禁止** `ResolveVoiceWxID` deviceNo 反查替代登录
- [x] 2.8 补充中文注释：跨服务调用、TTL 边界、额度与限流错误语义

### 3. gateway-app-server：App 对外 WS 入口（MUST）

- [x] 3.1 `internal/controller/ws_route_proxy.go`：将 `/voice/clinic/ws` 加入 `voiceWSProxyPaths`（与 `/voice/chat/ws`、`/voice/asr/ws` 共用 `installVoiceWSProxyMiddleware` 与 `VOICE_WS_PROXY_URL`）
- [x] 3.2 `gateway_app_auth_exempt.go`：将 `/voice/clinic/ws` 加入 `gatewayAppAuthExemptExactGET`（Bearer 豁免，同 `/voice/asr/ws`）
- [x] 3.3 确认 `gateway_app_register.go` → `RegisterGatewayAppHTTP` 已调用 `installVoiceWSProxyMiddleware(s)`（无需新增 BindHandler，透传由中间件短路）
- [ ] 3.4 手工验证：App 经 gateway-app `wss://{apiBaseUrl}/voice/clinic/ws` Upgrade 成功；无 Bearer 可进入透传链；**不得**要求直连 voice-service

### 3b. gateway-service：同路径透传同步（SHOULD）

- [x] 3b.1 `register.go` 侧 `voiceWSProxyPaths` 同步加入 `/voice/clinic/ws`（管理/通用网关与 App 网关路径一致）

### 4. 合规与文档

- [x] 4.1 更新 `resource/public/privacy-policy.html`：修正 DashScope/DeepSeek 表述；披露胖宝 7 天摘要与 thinking 展示；更新生效日期
- [x] 4.2 核对 Redis 策略：proposal/design 已记录「负责人已确认」`voice:clinic:*` 键；实现与 design 一致
- [x] 4.3 App API usage 统计：`/voice/clinic/ws` 为 WebSocket，**不统计**（结构性 skip，无需 maintenance_skip 变更）

### 5. 验收（go_ai_talk）

- [ ] 5.1 手工验证：clinic WS 首帧 `auth`（wxId 绑定）、流式 thinking+answer、12h session 不续期、摘要 watermark 刷新、40301/40302/42901
- [ ] 5.2 手工验证：gateway-app 透传与无 Bearer Upgrade；Flutter/env 使用 gateway-app 主机而非 voice-service；与 `/voice/chat/ws` 并存不互踢；同 wxId 不同 deviceNo 的 voice ball 不受影响
- [x] 5.3 确认无新增 `*_test.go`、无 background ticker

---

## flutter_ai_talk

### 1. 入口与环境

- [x] 1.1 `home_immersive_header.dart`：**保留** 趋势入口（`Icons.insights` → `/trends`），**新增** 胖宝入口（`Icons.pets` → `/pangbao`）与之并列
- [x] 1.2 环境配置：增加 `wsClinicUrl` + `wsClinicUrlEffective`（默认由 `apiBaseUrl` 推导 `wss://{host}/voice/clinic/ws`，对齐 `wsVoiceAsrUrlEffective`；**MUST NOT** 指向 voice-service 内网地址）

### 2. 胖宝页与 WebSocket

- [x] 2.1 新建 `pangbao_ai_screen.dart`：文本输入、消息列表、thinking 区、answer 区、免责声明
- [x] 2.2 新建 `clinic_ws_client.dart`：连接后首帧 `type=auth`（`accessToken` + `deviceNo`）；`auth_ok` 后发送 `question`、解析下行帧；生命周期对齐 UCG chat（后台 disconnect）
- [x] 2.3 独立 consent：`pangbao_ai_consent_v1`，首次进入展示同意，未同意禁止提问

### 3. Thinking UI 与额度

- [x] 3.1 Thinking 展示：默认最多 5 行可见、流式 auto-scroll、超 5 行折叠、tap 展开/局部滚动、用户 pin scroll 停止跟随
- [x] 3.2 每条 AI 回答展示「本回答仅供参考，不能替代医生诊断」
- [x] 3.3 `ai_quota_models` 增加 `clinicAi`；40302 弹框「本月额度已用完」、40301 引导登录

### 4. 验收（flutter_ai_talk）

- [ ] 4.1 端到端：首页进胖宝 → consent → WS `auth` → 提问 → 流式 thinking/answer → 免责文案
- [ ] 4.2 后台/前台 WS 断开与重连（重连须重新 `auth`）；未登录 40301、额度用尽 40302 错误 UI
