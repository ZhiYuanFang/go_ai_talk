## Why

胖宝诊疗 WS 抓包显示 thinking 结束后无 `answer_delta`，用户只见思考块、不见正文。根因：`parseOpenAIStreamDelta` 按分片解析，reasoning 结束后的 `content` 分片进入 `content` 通道，clinic 未订阅 `OnContentDelta`。

## What Changes

- `aimodel.invokeStreamHTTP`：当 `ThinkingEnabled` 且流中已出现过 reasoning 时，将后续 `content` 分片路由到 `answerBuf` 与 `OnAnswerDelta`。
- 闲聊等 `ThinkingEnabled=false` 路径不变，仍走 `OnContentDelta`。

## Capabilities

### Modified Capabilities

- `pangbao-ai-clinic`：thinking 流结束后 MUST 下发 `answer_delta` 与非空 `answer_done.answer`（上游有正文时）。

## Impact

- `internal/services/aimodel/client.go`、`provider.go`（注释）
- 无 WS 协议变更
