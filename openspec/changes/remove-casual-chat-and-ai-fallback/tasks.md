## 1. 删除闲聊模式枚举与模式管理代码

- [ ] 1.1 删除 chat_enums.go 中 ChatModeCasual 枚举值及相关解析逻辑
- [ ] 1.2 删除 voice_chat_understanding.go 中模式管理函数（detectChatModeByTranscript、isModeSwitchCommand、isModeQueryCommand、chatModeDisplayName、setDeviceChatMode、getDeviceChatMode、resolveChatMode）
- [ ] 1.3 删除 VoiceService 结构体中 modeMu 和 deviceModes 字段
- [ ] 1.4 删除 NewVoiceService 中 deviceModes 初始化逻辑
- [ ] 1.5 编译验证无编译错误

## 2. 删除闲聊流式回复与直接回复代码

- [ ] 2.1 删除 voice_chat_deepseek.go 中 callDeepSeekDirectReply 函数
- [ ] 2.2 删除 voice_chat_deepseek.go 中 streamCasualReplyWithBaiduTTS 函数
- [ ] 2.3 删除 voice_chat.go 中 StreamCasualReplyWithBaiduTTS 对外方法
- [ ] 2.4 删除 runtime_contracts.go 中 VoiceContract 接口的 StreamCasualReplyWithBaiduTTS 方法
- [ ] 2.5 编译验证无编译错误

## 3. 简化 chatWithResult 与 chatResult 结构

- [ ] 3.1 删除 chatResult 结构体中 Mode 和 NeedCasualStream 字段
- [ ] 3.2 删除 chatWithResult 函数的 directCasual 参数
- [ ] 3.3 删除 chatWithResult 中与闲聊模式相关的分支逻辑
- [ ] 3.4 编译验证无编译错误

## 4. 简化 HandleTranscriptForStreaming 接口签名

- [ ] 4.1 修改 voice_chat.go 中 HandleTranscriptForStreaming 函数签名，移除 mode 和 needCasualStream 返回值
- [ ] 4.2 修改 voice_chat.go 中 HandleTranscriptChatOnly 函数签名（如适用）
- [ ] 4.3 修改 runtime_contracts.go 中 VoiceContract 接口的对应方法签名
- [ ] 4.4 修改 voice_ws.go 中对 HandleTranscriptForStreaming 的调用，移除 mode 和 casualFlow 变量
- [ ] 4.5 删除 voice_ws.go 中 casualFlow 整个分支（第183-227行）
- [ ] 4.6 编译验证无编译错误

## 5. 删除 AI 能力兜底代码（意图分析/成长建议/历史问答）

- [ ] 5.1 删除 callDeepSeekUnifiedIntent 中 DeepSeek 直连兜底逻辑（第543-605行），Python 失败时直接返回错误
- [ ] 5.2 修改 chatWithResult 中 callDeepSeekUnifiedIntent 失败时的处理，返回固定降级提示语 "AI 服务暂时不可用，请稍后再试"
- [ ] 5.3 删除 callDeepSeekGrowthSuggestion 中 LLM 直连兜底逻辑（第247-288行），Python 失败时返回固定降级提示语
- [ ] 5.4 删除 callDeepSeekHistoryReply 中 DeepSeek 直连兜底逻辑（第483-507行），Python 失败时返回固定降级提示语
- [ ] 5.5 编译验证无编译错误

## 6. 删除诊疗流式兜底代码

- [ ] 6.1 删除 streamClinicLLMHeld 中 LLM 直连兜底逻辑（clinic_llm.go 第42-60行），Python 失败时直接返回错误
- [ ] 6.2 删除 buildClinicLLMMessages 函数（如确认无其他调用方）
- [ ] 6.3 上层调用 Python 失败时返回固定降级提示语
- [ ] 6.4 编译验证无编译错误

## 7. 清理死代码与辅助函数

- [ ] 7.1 删除 voice_chat_understanding.go 中 callDeepSeekActionExtract 函数（死代码，未被调用）
- [ ] 7.2 删除 handleIntentGeneral 函数（死代码 + 兜底，第1024-1071行）
- [ ] 7.3 确认 callDeepSeekRaw 保留（被 callDeepSeekEntityExtract 依赖）
- [ ] 7.4 确认 callDeepSeekEntityExtract 保留（业务兜底，非 AI 兜底）
- [ ] 7.5 清理已无引用的辅助函数和常量（如 parseGeneralChatResult 等，需逐个确认）
- [ ] 7.6 编译验证无编译错误

## 8. 全局清理与验证

- [ ] 8.1 全局搜索 ChatModeCasual、casual、闲聊 等关键词，确认无遗漏
- [ ] 8.2 全局搜索 DeepSeek 兜底相关日志（"回退到 DeepSeek"、"回退到 LLM"），确认已全部清理
- [ ] 8.3 运行 go build 验证全项目编译通过
- [ ] 8.4 运行 go vet 检查无告警
- [ ] 8.5 手工验证主流对话场景正常（意图识别、动作记录、成长建议、历史问答）
