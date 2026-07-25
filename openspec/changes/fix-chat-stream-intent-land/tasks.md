## 1. 抽取公共落库路径

- [x] 1.1 在 `voice_chat_understanding.go` 从 `chatWithResult` 抽取「已有 `deepSeekUnifiedIntent` → NeedConfirm / conversation / Action + `handleUnifiedIntentAction`」为可复用函数（如 `applyUnifiedIntentResult`），含详细中文注释
- [x] 1.2 将 pending confirm / pending child / 文本规范化等 preamble 整理为可被流式入口复用的逻辑（抽取或清晰内联调用），保证与现网优先级一致
- [x] 1.3 重构 `chatWithResult`：常规路径仍 `callDeepSeekUnifiedIntent`，成功后改为调用公共 apply 函数；行为与抽取前等价

## 2. 流式入口直接落地

- [x] 2.1 改写 `HandleTranscriptForIntentStream`：先走 preamble（pending confirm / pending child）；仅在需要新意图时调用一次 `callDeepSeekUnifiedIntentStream`
- [x] 2.2 Stream 成功后用 `mapPythonRespToIntent(streamRes.Result)` 映射，再调公共 apply；**删除**对 `chatWithResult` 的二次意图调用
- [x] 2.3 Stream 失败或 `Result == nil` 时返回错误/降级话术，**禁止**回退 `AnalyzeIntent` / `callDeepSeekUnifiedIntent`
- [x] 2.4 对齐 quota：guard → 单次 Stream 成功 → 单次 `consumeVoiceAIQuotaOnSuccess`；pending confirm 恢复路径不重复 consume

## 3. 行为矩阵与可观测性核对

- [x] 3.1 按 design 行为矩阵逐项核对：NeedConfirm / pending confirm / multi-event / `handleUnifiedIntentAction` / conversation
- [x] 3.2 为流式落地成功路径增加可观测日志（含 `intent_path=stream_land` 或等价标记、`deviceNo`、`target_type`、`action`、`need_confirm`）
- [x] 3.3 代码检索确认：`HandleTranscriptForIntentStream` 调用链上无 `callDeepSeekUnifiedIntent` / 非流式 `AnalyzeIntent`

## 4. 边界与编译

- [x] 4.1 确认修改仅限 `internal/services/voice`；无 gateway / Flutter / tip / feedback 改动；无他域 DAO import
- [x] 4.2 确认未新增测试文件；相关函数含中文业务注释
- [x] 4.3 编译通过（voice 相关包 / 工程可编译检查）
