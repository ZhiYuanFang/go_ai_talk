## Why

胖宝诊疗（clinic）当前仅将近 7 天喂养事件聚合摘要注入 DeepSeek system prompt，未携带宝宝出生日期与性别，导致 AI 无法按月龄与性别给出准确的育儿建议。语音球路径已通过 `DeviceProfile` HTTP 契约读取画像；clinic 应复用同一模式补齐上下文。

## What Changes

- 每次 clinic `question` 处理前，经 auth 锁定的 `deviceNo` 调用 `DeviceProfile().GetProfile` 获取 `birthday` 与 `sex`。
- 在 LLM system prompt 中，于 7 天喂养聚合摘要 JSON **之前**注入单行 JSON 宝宝信息块（A2 格式：`{"birthday":"…","gender":"…","age_months":N}`），与语音球画像字段风格对齐。
- 服务端根据 `birthday` 计算 `age_months`（整月差；未设置时为 0 或省略，见 design）。
- `DeviceProfile` HTTP 拉取失败时 **降级继续**：使用 `birthday="未设置"`、`gender="女"`（与 `loadDeviceProfile` 一致），记录 warning 日志，**不阻断** LLM 调用。
- 微调 `config.voice-service.yaml` 的 `aiClinic.systemPrompt`，说明需结合宝宝年龄、性别与喂养摘要回答。
- 更新隐私政策，披露胖宝诊疗会读取宝宝出生日期与性别用于适龄建议。
- **不改动**：WS 帧协议、Flutter 客户端、summary Redis 缓存结构/watermark 逻辑、原始 history 全量 dump。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `pangbao-ai-clinic`：新增「注入宝宝画像（birthday/sex/age_months）至 LLM system context」需求；明确 profile 契约失败降级语义。
- `app-legal-docs`：修订隐私政策中胖宝诊疗披露，补充出生日期与性别。

## Impact

- **代码**：`internal/services/voice/clinic_service.go`、`clinic_llm.go`；新增 `clinic_profile.go`（或等价包内函数）；可选微调 `clinic_config.go` / `manifest/config/config.voice-service.yaml`。
- **契约**：复用既有 `DeviceProfile` HTTP（device-service），voice-service 禁止直连 `dao.User`。
- **缓存**：不新增 Redis 键；profile 每轮实时拉取，不并入 `voice:clinic:summary:*`。
- **对外 API/WS**：无协议变更。
- **文档**：`resource/public/privacy-policy.html`。
