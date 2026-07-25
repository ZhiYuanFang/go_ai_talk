## ADDED Requirements

### Requirement: voice-service 注册 clinic/tip 反馈控制器
voice-service SHALL 在 HTTP 注册中 Bind `DeviceClinicFeedbackController`（或等价控制器），对外暴露已有 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback`。

#### Scenario: clinic feedback 可达
- **WHEN** 已鉴权客户端向 voice-service（经 gateway）发送 `POST /device/api/clinic/feedback`，Body 含合法 `answerId` 与 `feedback`（1 或 -1）
- **THEN** 控制器 SHALL 调用 `PythonAIClient.ClinicFeedback` 并将 Python 结果映射为 `{code, message, data}` 响应

#### Scenario: tip feedback 可达
- **WHEN** 已鉴权客户端向 voice-service（经 gateway）发送 `POST /device/api/tip/feedback`，Body 含合法 `answerId` 与 `feedback`（1 或 -1）
- **THEN** 控制器 SHALL 调用 `PythonAIClient.TipFeedback` 并将 Python 结果映射为 `{code, message, data}` 响应

#### Scenario: 非法 feedback 值
- **WHEN** 请求 Body 中 `feedback` 不是 1 或 -1
- **THEN** 控制器 SHALL 返回业务错误（如 code=400）且 MUST NOT 调用 Python

### Requirement: gateway 反代 clinic/tip API 到 VOICE
gateway-app SHALL 将 `/device/api/clinic/*` 与 `/device/api/tip/*` 反代至 `VOICE_API_PROXY`（voice 域路由模式），MUST NOT 反代至 device-service。

#### Scenario: clinic/tip API 走 voice 上游
- **WHEN** gateway-app 收到匹配 `/device/api/clinic/*` 或 `/device/api/tip/*` 的请求且 voice 代理已启用
- **THEN** 请求 SHALL 经 voice 反代中间件转发到 voice-service
- **AND** MUST NOT 进入 device 域 `/device/app/api/feedback/*` 反代路径

#### Scenario: 与 device 用户反馈路径隔离
- **WHEN** 客户端请求 `/device/app/api/feedback/*`
- **THEN** 行为 SHALL 保持既有 device 域反代，不受本变更影响

### Requirement: feedback 接口不计入 usage 统计
负责人已确认：`POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback` MUST NOT 计入 gateway-app App API 使用统计。实现 SHALL 在 `maintenance_skip.go` 以精确 `METHOD + path` 排除二者。

#### Scenario: clinic feedback 被 skip
- **WHEN** gateway-app 处理 `POST /device/api/clinic/feedback`
- **THEN** usagestats SHALL 将其视为维护型 API 并跳过计数

#### Scenario: tip feedback 被 skip
- **WHEN** gateway-app 处理 `POST /device/api/tip/feedback`
- **THEN** usagestats SHALL 将其视为维护型 API 并跳过计数

#### Scenario: tip generate 不在本包改统计策略
- **WHEN** 评估 `POST /device/tip/generate` 的 usage 策略
- **THEN** 本变更 MUST NOT 将其写入 `maintenance_skip.go`（统计属包 B）

### Requirement: 反馈路径与鉴权保持既有契约
反馈 HTTP 路径 MUST 保持为 `/device/api/clinic/feedback` 与 `/device/api/tip/feedback`；经 gateway-app 暴露时 MUST 要求登录（不加入 Bearer exempt）。

#### Scenario: 路径不变
- **WHEN** 查阅 `api/v1` 的 `g.Meta` 或对外文档
- **THEN** clinic/tip 反馈 path 仍为上述两路径，无破坏性改名

#### Scenario: 未登录被拒
- **WHEN** 无有效 Bearer 的客户端请求上述反馈 POST
- **THEN** gateway-app SHALL 按既有鉴权规则拒绝（401），不得因本变更加入白名单
