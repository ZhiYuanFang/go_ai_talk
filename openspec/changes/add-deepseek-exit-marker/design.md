# Design: DeepSeek exit-marker

## Overview

检测到用户意图为退出对话时，后端需要返回一个机器可识别的退出标记（例如 `{"exit": true}`），
前端收到该标记后应停止播放任何合成音频并进入待唤醒模式。设计必须最小侵入、向后兼容。

## Where to detect

- 在 `internal/service/voice_chat.go` 中的 DeepSeek 响应解析之后，统一检测是否为退出意图。

## API surface

- 对外 HTTP/WebSocket 响应增加可选字段 `exit`（布尔）。
- 当 `exit == true` 时，`reply` 可能为空且不会触发 TTS。

## Backwards compatibility

- 若后端未检测到 `exit` 或未部署该功能，行为保持不变（无 `exit` 字段或 `exit: false`）。
