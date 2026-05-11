## ADDED Requirements

### Requirement: App 网关进程独立运行

系统 SHALL 提供名为 gateway-app-server 的独立 HTTP 服务进程，具备与现有 gateway 相当的静态资源与领域反向代理能力，并额外承载 App 鉴权、令牌、版本检查与历史 WebSocket。

#### Scenario: 进程启动与配置隔离

- **WHEN** 使用 gateway-app-server 专用配置文件启动进程
- **THEN** 服务 SHALL 仅加载该进程所需的数据库分组（含 ai_voice_app）与下游 URL 配置，且 SHALL NOT 将 voiceChat 等业务配置错误合并到错误进程的权威配置源中（遵循仓库既有配置边界约定）

### Requirement: Bearer 鉴权与内部头注入

系统 SHALL 对除白名单外的受保护 HTTP 路径校验 `Authorization: Bearer <access_token>`，其中 `access_token` **MUST** 为符合 RFC 7519 的 **JWT**；系统 SHALL 先验证 JWT 签名与 `exp`，再从 **标准 JWT claim `sub`** 解析出大于 0 的整数 `wx.id`，通过 device-service 契约获取对应 `wxCode` 并写入转发请求头 `X-Internal-Wx-Code` 后再进入反向代理链。

#### Scenario: 鉴权通过并代理

- **WHEN** 客户端请求受保护路径且 Bearer 为合法未过期的 JWT、`sub` 对应 wx.id 在 device 侧存在
- **THEN** 网关 SHALL 在发往 device-service 或 history-service 的代理请求上设置 `X-Internal-Wx-Code`，且 SHALL NOT 依赖修改原始 HTTP body 来传递 wxCode

#### Scenario: 鉴权失败

- **WHEN** Bearer 缺失、非 JWT、JWT 签名校验失败、已过期或 `sub` 无法解析为有效 wx 行
- **THEN** 网关 SHALL 拒绝请求并返回明确错误响应，且 SHALL NOT 设置 `X-Internal-Wx-Code`

### Requirement: 登录与令牌仅由 gateway-app 签发

系统 SHALL 在 gateway-app-server 上暴露 `POST /device/app/api/login`，其通过 HTTP 调用 device-service 的 `POST /device/wx/api/login` 获取业务字段后签发 access_token 与 refresh_token；其中 **access_token SHALL 为纯 JWT**，其载荷 **MUST** 包含标准 claim **`sub`** 且其值等于 wx 表主键 id（与 device 返回的 wxId 一致），**MUST** 包含 **`iat`** 与 **`exp`**；**refresh_token SHALL NOT** 为 JWT，SHALL 为高熵不透明串并与 Redis 会话绑定以便刷新与吊销。

#### Scenario: 登录成功

- **WHEN** App 调用 gateway-app 的登录接口且 device 返回有效业务结果
- **THEN** 响应 SHALL 包含 access_token 与 refresh_token（及 device 返回的业务字段），且 access_token SHALL 可被验证为结构正确的 JWT，且 device-service SHALL NOT 在 `POST /device/wx/api/login` 响应中返回 JWT 形式的 access_token

### Requirement: 刷新令牌接口

系统 SHALL 在 gateway-app-server 提供刷新 access 的 HTTP 接口（路径位于 `/device/app/api/` 前缀下），使用 Redis 校验 refresh 后签发新的 **JWT** 形态 access_token（`sub`/`iat`/`exp` 规则与登录接口一致），并可按产品策略旋转 refresh_token。

#### Scenario: 刷新成功

- **WHEN** 客户端提交有效 refresh_token
- **THEN** 系统 SHALL 返回新的 access_token 且该 token SHALL 为合法 JWT，且旧 refresh 的处理策略（保留至过期或立即失效）SHALL 与设计文档一致并在实现中单一实现

### Requirement: 版本检查 API

系统 SHALL 在 gateway-app-server 提供 `GET /device/app/api/version/check`，从查询参数读取 `currentVersion`，读取 ai_voice_app.version 表（或经缓存的等价数据）并返回 needUpdate、latestVersion、releaseNotes、downloadUrl、forceUpdate。

#### Scenario: 返回版本信息

- **WHEN** 客户端携带合法 currentVersion 调用版本检查接口
- **THEN** 响应 SHALL 包含布尔 needUpdate 及 latestVersion、releaseNotes、downloadUrl、forceUpdate 字段，且 MAY 使用 Redis 缓存版本行以降低数据库压力

### Requirement: 历史 WebSocket 与首帧认证

系统 SHALL 在 gateway-app-server 提供 WebSocket 端点；连接建立后首条文本帧 MUST 为 JSON，包含 `type` 为 `auth`、`access_token`（snake_case 键名，值为 **JWT 字符串**）与 `device_no`；服务端 MUST 按与 HTTP Bearer 相同的规则校验 JWT 后，再校验 `sub` 对应 wx 身份与该 device_no 的绑定关系，通过后才将连接注册到按 device_no 分组的推送集合。

#### Scenario: 认证成功并订阅

- **WHEN** 客户端发送合法 auth 帧且 access_token 为有效 JWT、device_no 与该 token 身份匹配
- **THEN** 连接 SHALL 保持打开并能够接收后续由 Redis 通知触发的历史变更消息

#### Scenario: 认证失败

- **WHEN** auth 帧缺失、字段不合法或 device_no 与身份不匹配
- **THEN** 服务端 SHALL 拒绝订阅（关闭连接或发送错误文本帧）且 SHALL NOT 将该连接加入任何 device_no 推送组

### Requirement: Redis Pub/Sub 消费与下行

系统 SHALL 在 gateway-app-server 进程内维护对约定 Redis channel 的订阅；当收到 history-service 发布的消息时，SHALL 向所有已认证且匹配 `device_no` 的 WebSocket 连接推送 JSON 业务消息。

#### Scenario: 收到发布并推送

- **WHEN** Redis 收到一条包含已知 device_no 与历史载荷的合法通知
- **THEN** 网关 SHALL 向该 device_no 下已注册且仍存活的连接广播该消息体

### Requirement: 鉴权白名单

系统 SHALL 对 `POST /device/wx/api/login`（若经网关暴露）、gateway-app 的登录与刷新接口、版本检查（若产品要求公开）、WebSocket 握手路径等无需 Bearer 的路径配置中间件白名单，使其不触发 Bearer 解析失败。

#### Scenario: 无令牌访问登录

- **WHEN** 客户端无 Authorization 头调用白名单内的登录接口
- **THEN** 请求 SHALL 进入对应处理器且 SHALL NOT 被 Bearer 中间件以「未授权」拦截
