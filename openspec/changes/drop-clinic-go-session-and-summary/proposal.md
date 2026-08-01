## Why

胖宝 Clinic WS（产品侧称「智能陪伴」）的多轮对话已由 Flutter 本地持久化展示，并由 Python `companion_session`（按 `device_no`）维护 agent 上下文；Go 侧仍维护 Redis `voice:clinic:session:*`、在 `auth_ok` 后下发 `session_sync`，并在每轮 question 前计算/缓存近 7 天喂养摘要与宝宝画像——这些数据实际已不注入 `ClinicStream`（仅传 `question`/`device_no`/`model`），造成三端重复、额外 Redis 与 history 拉取，且与「换机/清本地不补历史」产品决策冲突。

## What Changes

- **BREAKING（相对旧 Clinic WS 协议）**：`auth_ok` 之后 **MUST NOT** 再下发 `type=session_sync`；客户端在 `auth_ok` 后即可发送 `question`/`cancel`。
- **删除** Go Redis 对话会话：不再读写 `voice:clinic:session:{wxId}`，不再 `appendClinicTurn`；换机/清本地 **不** 由 Go 补历史对话。
- **删除** 近 7 天喂养事件聚合摘要路径：`ensureClinicSummary` / `clinic_summary.go` / `voice:clinic:summary:*` 及 question 前对 `DeviceHistory` 的摘要拉取。
- **删除** 未使用的 `loadClinicBabyProfile`（`clinic_profile.go`）及对 `streamClinicLLM*` 的死参（`baby`/`summary`/`prior`）。
- **保留**：Clinic WS 鉴权、`clinic_ai` 额度、per-wxId 限流（`voice:clinic:rate:*`）、LaneClinic 闸门、Python `ClinicStream` 代理、`answer_*` / `turn_cancelled` / `error` 帧。
- **命名**：本仓路径/类型仍用 clinic，不强制改名为 companion。
- **范围**：仅 `go_ai_talk`；不改 Flutter / Python 仓（Flutter 已忽略 `session_sync`）。
- **文档**：同步修订 `privacy-policy.html` 中「Go/系统读取近 7 天喂养聚合摘要」等与实现不符的表述（摘要职责已不在 Go）。

## Capabilities

### New Capabilities

- （无）行为收敛到既有 `pangbao-ai-clinic` 能力上的删减与协议收窄。

### Modified Capabilities

- `pangbao-ai-clinic`：取消 `session_sync` 与 Redis 会话；取消 Go 侧 7 天喂养摘要与宝宝画像注入；Clinic WS 在 `auth_ok` 后直接进入问答循环；Go 仅作鉴权/额度/限流/Python 流转发。

## Impact

- **后端（本仓）**：`internal/controller/voice_clinic_ws.go`；`internal/services/voice/clinic_service.go`、`clinic_llm.go`、`clinic_session.go`（删）、`clinic_summary.go`（删）、`clinic_profile.go`（删）、`clinic_config.go`；`internal/platform/cachekit/keys_voice.go`；`manifest/config/config.voice-service.yaml`（去掉 session/summary TTL 与摘要向 systemPrompt）；`resource/public/privacy-policy.html`；对照基线 `openspec/specs/v2.0.24/spec.md` 中 `pangbao-ai-clinic`。
- **Flutter / Python**：本变更不改代码；多轮与 UI 历史分别依赖既有 Python Redis 会话与 Flutter 本地 store。
- **Redis**：停止写入 `voice:clinic:session:*`、`voice:clinic:summary:*`（旧键随 TTL 自然过期即可，无需迁移脚本）；保留 `voice:clinic:rate:*`。
- **网关 / usage**：无新增 App HTTP 路由；不涉及 usage denylist。
- **风险**：旧客户端若强依赖非空 `session_sync` 重建列表会失效——现网 Flutter 已忽略该帧，可接受。
