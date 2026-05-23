## 1. 听写 WS 截句逻辑（方案 A）

- [x] 1.1 在 `voice_asr_ws.go` 移除 `silence` / `noFirstSTT` / `stTimeout` 与 `tryAutoCommitWhenNoASRCallback` 及关联状态字段
- [x] 1.2 将 `CreateStreamASRSession` 的 `onFinal` 改为 no-op（不 `emitAsrFinal`、不 `resetStreamASRUntilNextValid`）
- [x] 1.3 确认仅 `commit` / `end` 调用 `runAsrFinalize`，且 `source` 仅为 `client` 或 `end`
- [x] 1.4 补充中文注释：听写线截句权在前端，与 chat WS 差异

## 2. 文档

- [x] 2.1 更新 `README.MD` 听写 WS：`asr_final` 必须由 `commit`（或 `end`）触发；删除 silence/auto_commit/asr_callback 说明
- [x] 2.2 给前端联调一句：松手/完成须发 `{"type":"commit"}`

## 3. 验证

- [x] 3.1 `go build ./...` 通过
- [x] 3.2 对照本变更 spec 场景：静音不发 `asr_final`、commit 才发 `asr_final`、日志无 ASR 建连风暴
