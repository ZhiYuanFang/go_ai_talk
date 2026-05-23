## 1. 听写 WS onFinal 转发

- [x] 1.1 在 `voice_asr_ws.go` 将 `CreateStreamASRSession` 的 `onFinal` 改为：非空且 `text != lastPartialText` 时更新 `lastPartialText`、`latestTranscript`（直接赋值）并 `emitAsrPartial`
- [x] 1.2 确认 `onFinal` 不调用 `emitAsrFinal`、`resetStreamASRUntilNextValid` 或 `runAsrFinalize`
- [x] 1.3 补充中文注释：引擎 final 仅作预览纠正，业务定稿仍靠 `commit`/`end`

## 2. 文档

- [x] 2.1 更新 `README.MD` 听写 WS：说明引擎可能在说话中推送纠正性 `asr_partial`；`asr_final` 仍仅 `commit`/`end`

## 3. 验证

- [x] 3.1 `go build ./...` 通过
- [x] 3.2 对照 spec：onFinal 下发 partial、不下发 asr_final、不关 ASR
