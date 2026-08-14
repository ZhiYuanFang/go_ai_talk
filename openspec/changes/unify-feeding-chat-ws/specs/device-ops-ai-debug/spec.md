## ADDED Requirements

### Requirement: 运维页支持 App 用户名登录

设备数据运维单设备页（`/device/admin/history/{deviceNo}` 所载 `history.html`）MUST 提供 App 用户名密码登录入口，调用既有 `POST /device/app/api/username_login`，并将返回的 `accessToken` 存于与 Admin JWT **不同**的存储键。Admin JWT 继续用于既有 admin API；tip 与 care-alert 调试请求 MUST 使用 App `accessToken` 作为 Bearer。

#### Scenario: 未 App 登录不可调 tip/care-alert 调试

- **WHEN** 运维未完成 App 用户名登录即点击小贴士生成或值得留意拉取
- **THEN** 页面 MUST 提示需先登录，MUST NOT 使用 Admin JWT 调用上述 App 接口

#### Scenario: Admin 与 App token 分钥

- **WHEN** 运维同时持有 Admin JWT 与 App accessToken
- **THEN** 历史 CRUD MUST 仍带 Admin Bearer；tip/care-alert MUST 带 App Bearer

### Requirement: 运维文字对话调试走 chat WS 文模式

运维页 MUST 提供文字对话调试面板：连接 `/voice/chat/ws`，使用 text 输入与 text 输出模态，展示服务端 `thinking_delta` 与 `answer`。MUST NOT 为此面板新增 gateway 鉴权豁免或 Admin 伪造身份接口。

#### Scenario: 发送文本可见思考与回答

- **WHEN** 运维在已 start（文模式）的连接上发送非空文本
- **THEN** 面板 MUST 展示思考增量与最终 answer（允许忽略音频帧）

### Requirement: 运维小贴士调试走正式 tip SSE

运维页 MUST 通过 `POST /device/tip/generate`（body 含当前 `deviceNo` 及可选 event 字段）发起 SSE，展示 `thinking`/`answer`/`done`（或等价事件）。MUST NOT 新增 tip 的 Admin 旁路或鉴权豁免。重复点击生成即视为强刷（服务端无日缓存门禁时仍走正式额度语义）。

#### Scenario: tip SSE 展示思考过程

- **WHEN** 运维已 App 登录并对当前设备触发小贴士生成
- **THEN** 页面 MUST 以流式方式展示思考与最终建议内容

### Requirement: 运维值得留意调试走正式 care-alert API

运维页 MUST 通过 `GET /device/api/care-alert/daily` 拉取当日列表并展示；MUST 提供强刷操作，请求带 force 语义（见 `care-alert-force-refresh`）。MUST NOT 使用 Admin JWT 或伪造 `X-Internal-Wx-Id` 调用该接口。

#### Scenario: 拉取当日值得留意

- **WHEN** 运维已 App 登录并点击拉取值得留意
- **THEN** 页面 MUST 展示返回的 `day` 与 `items`（可空列表）

#### Scenario: 强刷按钮带 force

- **WHEN** 运维点击值得留意强刷
- **THEN** 页面 MUST 使用带 force 的 daily 请求（仍携带 App Bearer）
