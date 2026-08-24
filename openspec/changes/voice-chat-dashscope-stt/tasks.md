## 1. 配置与类型

- [x] 1.1 在 `VoiceChatConfig` / `definitions.go` 增加 `STTChat`、`STTDictation` 结构（或 `sttChat`/`sttDictation` JSON 字段），含 provider、model、streamEndpoint、workspaceId、apiKey、speechNoiseThreshold 等
- [x] 1.2 更新 `loadVoiceConfig`：解析 `sttChat`/`sttDictation`；未配置 `sttDictation` 时 fallback 至现有 `stt` 块（听写行为不变）
- [x] 1.3 更新 `manifest/config/voice-chat.shared.yaml`：`sttChat` 设为 dashscope + `qwen-audio-3.0-asr-flash-streaming`；保留/迁移 `stt` 为听写百度配置
- [x] 1.4 更新 `manifest/docker/.env.example`：增加 `VOICE_DASHSCOPE_API_KEY`、`DASHSCOPE_WORKSPACE_ID` 说明

## 2. DashScope 流式 STT 实现

- [x] 2.1 新增 `internal/services/voice/stt_stream_dashscope.go`：实现 `dashscopeStreamASRSession`（run-task、二进制 PCM、result-generated 解析、finish-task、readLoop）
- [x] 2.2 实现 API Key 解析链：`sttChat.apiKey` → `VOICE_DASHSCOPE_API_KEY` → `UCG_DASHSCOPE_API_KEY`
- [x] 2.3 实现 WebSocket URL 构建：`workspaceId` + 默认 `cn-beijing` 端点；握手 `Authorization: Bearer`
- [x] 2.4 映射 partial/final 回调与 `Commit` 等待语义，对齐 `stt_stream_baidu.go` 边界（WAV 头剥离、空音频、超时）
- [x] 2.5 增加中文业务注释（文件/方法/关键分支）

## 3. Profile 分流与契约

- [x] 3.1 定义 `STTProfile` 类型（`chat` | `dictation`），扩展 `StreamASRSession` 创建入口签名
- [x] 3.2 更新 `internal/services/contracts/runtime_contracts.go` 的 `VoiceContract.CreateStreamASRSession` 增加 profile 参数
- [x] 3.3 更新 `voice_chat.go` 的 `CreateStreamASRSession`：按 profile 分发 dashscope / baidu；更新 `contracts_aliases` 若有必要
- [x] 3.4 `voice_ws.go` 传入 `profile=chat`；`voice_asr_ws.go` 传入 `profile=dictation`

## 4. 非 WS 与降级（可选）

- [x] 4.1 评估 `transcribe` / `HandleWithDialogue` 整段 STT：chat 场景是否需 DashScope 或暂保留百度（design 非目标路径可标记 defer）
- [x] 4.2 （可选）实现 `sttChat.fallbackProvider=baidu` 建连失败降级

## 5. 部署与验收

- [x] 5.1 test 环境配置 DashScope 凭证与 workspaceId，验证 `/voice/chat/ws` 远场样例 asr_partial/asr_final
- [x] 5.2 验证 `/voice/asr/ws` 听写仍走百度且无协议回归
- [x] 5.3 检查日志含 profile/provider/model；avgAbs 偏低告警仍生效
- [x] 5.4 更新 `docs/runbooks/release-deploy-and-run.md`（若涉及 voice 环境变量）说明对话 STT 切换与回滚方式
- [x] 5.5 proposal 自检：无新 Redis 键、无 gateway 新路由、WebSocket 不计 usage（维持现状）

## 6. 评审清单

- [x] 6.1 确认 `openspec/specs/v3.0.0` 中 `voice-realtime-asr-ws` 听写要求仍满足（百度未改）
- [x] 6.2 确认 `internal/services/voice` 无跨域 DAO 直查
- [x] 6.3 确认主配置 `config.yaml` 未回流 voice STT 专属字段
