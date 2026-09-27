## 1. 配置与降级

- [x] 1.1 将 `manifest/config/voice-chat.shared.yaml` 的 `sttDictation` 改为 dashscope（字段对齐 `sttChat`）
- [x] 1.2 修正 `CreateStreamASRSession`：DashScope→百度降级使用 legacy `stt`，不用 `sttDictation`

## 2. 注释与文档

- [x] 2.1 更新 `voice_chat.go` / `definitions.go` / `voice_asr_ws.go` 中「听写=百度」相关中文注释
- [x] 2.2 更新 `docs/runbooks/release-deploy-and-run.md`：听写亦依赖 DashScope Workspace/Key；回滚说明含 `sttDictation`

## 3. 校验

- [x] 3.1 确认听写仍走 `STTProfileDictation` + `CreateStreamASRSession`，WS 协议未改
- [x] 3.2 `openspec validate voice-dictation-dashscope-stt --strict`（若 CLI 支持）或等价自检
