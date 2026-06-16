## Why

竞品喂养类 App 已提供「结合喂养记录的 AI 问答」能力；用户若在本 App 提问，需反复口述近期喂养数据，体验差、转化弱。新增用户向 **「胖宝」AI 诊室**（内部 `clinic` / `clinic_ai`），在提问前自动注入近 **7 天喂养事件聚合摘要**，让用户无需重复陈述即可获 DeepSeek 流式回答（含思考过程展示），以差异化吸引竞品用户。

## What Changes

- **新增** 胖宝 AI 诊室 WebSocket，**App 对外入口**为 **gateway-app-server** 路径 `GET /voice/clinic/ws`（与 `/voice/chat/ws`、`/voice/asr/ws` 同域、同 `apiBaseUrl` 主机）；gateway-app **MUST** 注册 WS 透传（`ws_route_proxy.go` → `voiceWSProxyPaths`）与 Bearer 鉴权豁免；**voice-service** 承载 `/voice/clinic/ws` 业务 handler（流式 DeepSeek `deepseek-v4-pro`，thinking 开启），帧类型 `thinking_delta` / `answer_delta` / `answer_done` / `error`；**不**注册 `VoiceWSManager`；**无 TTS**；App **MUST NOT** 直连 voice-service。
- **新增** 胖宝 WS **wxId 主键绑定**（与 `/voice/chat/ws` 的 deviceNo 主键、`VoiceWSManager` 互踢策略不同）：握手后首帧 `auth` 解析 JWT 得 `wxId>0`；会话/限流/额度均以 `wx.id` 为维度；`deviceNo` 仅用于 history 摘要拉取。
- **新增** 胖宝会话与限流 Redis 键空间 `voice:clinic:*`（session `voice:clinic:session:{wxId}` 12h 固定 TTL 自首问起算；rate `voice:clinic:rate:{wxId}`；summary 仍含 `{deviceNo}`）；**负责人已确认**引入 Redis。
- **新增** 7 天喂养历史 **聚合摘要**（按 event 统计 count/amount/duration），**懒刷新**：每问前对比 watermark 与 history 最后更新时间，过期则重算。
- **新增** `clinic_ai` 独立月度额度（默认 **30 次/月**，与 `voice_ai` 5 次分离）；device DB/API/Admin UI 第三字段；voice 在 LLM 前 check、成功后 consume。
- **新增** `config.voice-service.yaml` 独立 `aiClinic:` 块（含 **120s** LLM 超时，不复用 `voice-chat.shared.yaml` 20s）。
- **修改** **gateway-app-server**（App 网关，主入口）：`ws_route_proxy.go` 将 `/voice/clinic/ws` 加入 `voiceWSProxyPaths`；`gateway_app_auth_exempt.go` Bearer 豁免（策略同 `/voice/asr/ws`）；`RegisterGatewayAppHTTP` 已调用 `installVoiceWSProxyMiddleware` 挂载透传链。
- **修改** **gateway-service**（管理/通用网关，可选同路径）：同步 `voiceWSProxyPaths` 以保持一致（非 App 主入口）。
- **修改** 隐私政策 `privacy-policy.html`：修正 DashScope/DeepSeek 表述；披露胖宝 7 天摘要与 thinking 展示；**独立** consent 键 `pangbao_ai_consent_v1`（与首页喂养 AI consent 分离）。
- **跨仓 Flutter**（`flutter_ai_talk`）：首页沉浸式头部 **保留** 原 **趋势** 入口（`/trends`），**新增** **胖宝** 入口（`/pangbao`），二者并存、互不替换；新 `pangbao_ai_screen.dart`、`clinic_ws_client.dart`；`env wsClinicUrl` **MUST** 指向 gateway-app `wss://{apiBaseUrl host}/voice/clinic/ws`（**不得**配置 voice-service 内网地址）；thinking UI 交互规范；每条回复展示免责声明「本回答仅供参考，不能替代医生诊断」。
- **不新增** 测试文件、背景 ticker、TTS。

## Capabilities

### New Capabilities

- `pangbao-ai-clinic`：胖宝 AI 诊室 WS 协议（wxId 主键鉴权与绑定）、7 天摘要注入、DeepSeek 双流式、会话 TTL、独立限流与超时配置。
- `pangbao-ai-clinic-flutter`：Flutter 胖宝页、ClinicWsClient、thinking UI、consent 与免责声明展示（跨仓 `flutter_ai_talk`）。

### Modified Capabilities

- `ai-monthly-quota`：新增第三 feature `clinic_ai`（默认 30/月）；internal check/consume、App 读 API、Admin 全局/用户 override 均扩展第三字段。
- `gateway-ws-edge-proxy`：gateway-app-server **MUST** 注册 `/voice/clinic/ws` WS 透传与 Bearer 鉴权豁免；App 对外入口与 voice chat/ASR 同模式。
- `app-legal-docs`：隐私政策披露胖宝 7 天摘要、thinking 展示及模型供应商修正。

## Impact

- **go_ai_talk**
  - `internal/services/voice/`：clinic WS handler、摘要构建、DeepSeek streaming
  - `internal/services/device/`：ai_quota `clinic_ai` 列与 API
  - `internal/controller/ws_route_proxy.go`、`gateway_app_auth_exempt.go`、`gateway_app_register.go`（gateway-app-server WS 透传 + Bearer 豁免；voice-service 仅 handler）
  - `manifest/config/config.voice-service.yaml`：`aiClinic:` 块
  - `resource/public/privacy-policy.html`、device Admin UI
  - Redis 新键 `voice:clinic:session:{wxId}`、`voice:clinic:rate:{wxId}`、`voice:clinic:summary:{wxId}:{deviceNo}`（负责人已确认）
- **flutter_ai_talk**（`d:\work\flutter_ai_talk`）
  - `home_immersive_header.dart`、新胖宝页与 WS 客户端、`ai_quota_models`、`env wsClinicUrl`（默认由 `apiBaseUrl` 推导 `/voice/clinic/ws`，对齐 `wsVoiceAsrUrlEffective`）
- **依赖**：history-service HTTP 契约读取 7 天事件（禁止 voice 跨库直查）；device internal ai-quota
- **App API usage 统计**：`/voice/clinic/ws` 为 WebSocket，结构性 skip，**无需**计入 usage 统计确认
- **数据库**：device 域 `ai_quota` 相关表/配置增 `clinic_ai` 字段（进程 `device-service`，配置组 `database.default`）
