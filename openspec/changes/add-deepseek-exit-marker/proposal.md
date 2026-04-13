---
title: Add DeepSeek exit-marker support
change-id: add-deepseek-exit-marker
summary: |
  当 DeepSeek/STT 解析到用户意图为“退出对话”时，后端应返回一个明确的退出标记，
  以便前端终端进入待唤醒（待激活）状态而不是继续播放 TTS 或维持会话。
---

背景
----
当前语音流程：设备录音 -> STT -> DeepSeek（意图/聊天）-> 后端返回回复并合成 TTS。
当用户意图是“退出对话/结束会话”时，前端需要一个明确的信号来进入待唤醒状态，
而不是继续播放普通回复或保留会话上下文。

目标
----
- 在 DeepSeek 判定为“退出对话”时，后端返回退出标记给前端。
- 保持向后兼容，默认行为不变，只有在明确检测到退出意图时触发。
