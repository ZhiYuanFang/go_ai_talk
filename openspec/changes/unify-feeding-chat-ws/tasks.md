## 1. WS 四模态协议

- [x] 1.1 扩展 `/voice/chat/ws` 的 `start` 解析：`inputModality`/`outputModality`（默认 audio/audio）；音频元数据必填校验保持不变（文模式亦不放宽）
- [x] 1.2 实现 text 上行帧：在 `started` 后接受文本，跳过 ASR，进入既有 `HandleTranscriptForIntentStream`（Hardware 特权）
- [x] 1.3 按 `outputModality` 控制 TTS：`text` 时不下发 `audio_chunk`/`audio_end`；`audio` 时保持现网 TTS
- [x] 1.4 补充中文注释（协议分支、占位音频字段原因、与横屏默认兼容）

## 2. 删除外壳 B 与 internal text

- [x] 2.1 删除 history `Chat`/`ChatStream` 控制器方法与 `api/v1` 中对应 `g.Meta` 路由定义
- [x] 2.2 删除 `DelegateTextChat`/`DelegateTextChatStream` 及仅被其使用的 HTTP 辅助逻辑
- [x] 2.3 删除 voice `VoiceInternalTextChatCtrl`、internal text API 定义与 `register_voice_service` 绑定
- [x] 2.4 删除仅服务已删路径的 `POST /voice/text/chat`（若适用）及引用
- [x] 2.5 更新 gateway/voice-admin/Swagger 源中对已删路径的列举与说明

## 3. MCP 改挂 WS 文模式

- [x] 3.1 mcpbridge 喂养工具改为连接 `/voice/chat/ws`（text/text），提交 transcript，收取 `answer` 返回
- [x] 3.2 移除对 `DelegateTextChat` 的依赖与相关 history 委派配置说明
- [x] 3.3 更新 `docs/runbooks/release-deploy-and-run.md`（及 MCP 相关段落）中的接入描述与环境变量说明

## 4. Admin：VU free 删除与额度隐藏

- [x] 4.1 `ai-model-admin.html`：voiceUnderstanding 去掉 free 控件；标题可标「喂养默认智能体」；保存体不再提交 free
- [x] 4.2 voice/admin llm-lanes API 与 DTO：停止要求/持久化 VU free（忽略或拒绝入参）；正式模与并发字段保留
- [x] 4.3 `voice-admin.html` 额度区隐藏喂养（`voiceAi`）表单项；保留 clinic/care-alert（若有）
- [x] 4.4 清理喂养对话路径上对 VU free / `voice_ai` 额尽降级的依赖，确保 chat WS 仅用正式模+闸门

## 5. care-alert 强刷

- [x] 5.1 `GET /device/api/care-alert/daily` 增加 `force` 参数解析；force 为真时经 cachekit 删除当日 `CareAlertDailyKey` 后走既有生成路径
- [x] 5.2 保持无 force 时缓存命中语义不变；force 仍要求 `wxId>0`；打可观测日志
- [x] 5.3 API 定义/`g.Meta`/中文注释补充 force 语义（Additive，不改无 force 旧客户端）

## 6. 设备运维调试台

- [x] 6.1 `history.html`：App 用户名密码登录 UI + 独立 token 存储；与 Admin JWT 分钥
- [x] 6.2 文字对话面板：连接 `/voice/chat/ws` 文模式，展示 thinking/answer
- [x] 6.3 小贴士面板：App Bearer 调 `POST /device/tip/generate` SSE 展示流式结果（可重复触发）
- [x] 6.4 值得留意面板：App Bearer 拉取 daily；强刷按钮带 `force=1`
- [x] 6.5 登录后校验/提示 `deviceNo` 与 URL 设备一致性；补中文操作说明

## 7. 验收与文档

- [ ] 7.1 手工验收四种模态（至少文→文、音→音）；确认文模式缺 sampleRate 被拒
- [ ] 7.2 确认已删 HTTP 路径不可用；MCP 工具文模式可返回 answer
- [ ] 7.3 确认 Admin VU 无 free、额度页无喂养项；chat WS 在 voice_ai 用尽时仍可对话
- [ ] 7.4 确认运维台：App 登录后 tip/care-alert/WS 文对话可用；Admin token 无法代替 App 调 tip/care-alert；force 可重生值得留意
- [x] 7.5 确认无新增测试文件；无业务层 bypass cachekit；care-alert force 复用既有 key builder；无新 `*_DB_LINK` 需求；runbook/路径说明已更新
