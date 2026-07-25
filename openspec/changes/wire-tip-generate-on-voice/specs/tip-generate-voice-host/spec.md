## ADDED Requirements

### Requirement: TipCtrl 绑定到 voice-service

系统 SHALL 在 voice-service HTTP 注册路径中绑定 `NewTipCtrl()`（或等价），使 `POST /device/tip/generate` 由 voice 进程处理。

#### Scenario: voice 进程可路由 tip generate

- **WHEN** voice-service 启动并完成 `RegisterVoiceServiceHTTP`
- **THEN** `POST /device/tip/generate` MUST 由 `TipCtrl.Generate` 处理，且调用 `voice.TipStream` 写出 SSE

#### Scenario: 不绑定 feedback 控制器

- **WHEN** 本变更完成 voice 路由注册
- **THEN** 系统 MUST NOT Bind clinic tip feedback 相关控制器（留给包 C）

### Requirement: gateway 反代 /device/tip 到 voice

gateway-app SHALL 将 `/device/tip/*` 反向代理到 VOICE_API_PROXY（与现有 voice 域反代同一配置族）。

#### Scenario: App 请求经网关到达 voice

- **WHEN** 客户端向 gateway 发起 `POST /device/tip/generate`
- **THEN** 请求 MUST 被代理到 voice-service 对应路径，而非 history/device 本域处理

### Requirement: tip generate 需登录且计入 usage

`POST /device/tip/generate` SHALL 要求 Bearer 登录，且 MUST 计入 App API usage 统计。

#### Scenario: 不进入鉴权豁免

- **WHEN** 维护 `gateway_app_auth_exempt` 列表
- **THEN** `/device/tip/generate` MUST NOT 出现在 Bearer 豁免路径中

#### Scenario: 不进入 usage 维护排除

- **WHEN** 维护 `usagestats/maintenance_skip`
- **THEN** `/device/tip/generate` MUST NOT 被写入排除列表（即保持统计）

### Requirement: SSE 方言保持 Go 风格

`TipCtrl.Generate` SHALL 继续以 `event:` + data 纯文本写出 thinking/answer/done，done 的 data MUST 含 `answerId` 字段；最终 MUST 发送 `data: [DONE]`。

#### Scenario: 流式帧格式

- **WHEN** tip 流成功推送 thinking/answer/done
- **THEN** 每帧 MUST 含 `event: <name>` 与对应 `data:` 行，且 done 的 data 可解析出 `answerId`
