## MODIFIED Requirements

### Requirement: Flutter SHALL 提供胖宝诊疗入口与页面

App（`flutter_ai_talk`）MUST 在首页 `home_immersive_header.dart` **新增** **胖宝** 入口（`Icons.pets`，tooltip「胖宝」），跳转至新页面 `pangbao_ai_screen.dart`。诊疗页 AppBar 标题 MUST 为 **「胖宝诊疗」**。首页品牌标题与 tooltip **MUST NOT** 改为「胖宝诊疗」（保持「胖宝」）。页面 MUST 支持文本输入提问并通过 WebSocket 展示流式 thinking 与 answer。首页 **MUST 保留** 原 **趋势** 入口（`Icons.insights`，tooltip「趋势」→ `/trends`），胖宝入口 **MUST NOT** 替换或隐藏趋势入口。

#### Scenario: 从首页进入胖宝诊疗

- **WHEN** 用户在首页点击胖宝入口（tooltip「胖宝」）
- **THEN** App SHALL 导航至胖宝诊疗页面且 AppBar 显示「胖宝诊疗」

#### Scenario: 从首页进入趋势

- **WHEN** 用户在首页点击趋势入口
- **THEN** App SHALL 导航至趋势图表页面（`/trends`）

### Requirement: Flutter SHALL 使用 ClinicWsClient 经 gateway-app 连接 WebSocket

App MUST 实现 `clinic_ws_client.dart`。连接 URL MUST 使用 `wsClinicUrl` / `wsClinicUrlEffective`，默认由 `apiBaseUrl`（gateway-app-server 主机）推导为 `wss://{host}/voice/clinic/ws`（对齐 `wsVoiceAsrUrlEffective` 模式）。客户端 **MUST NOT** 配置或连接 voice-service 内网地址。连接成功后 MUST 先发送首帧 `type=auth`（`accessToken` + `deviceNo`，与 history WS / UCG chat 一致），收到 `auth_ok` 后方可发送 `question`。WS 生命周期 MUST 对齐 UCG chat：App 进入后台 MUST disconnect；回前台 MAY 重连且 MUST 重新 `auth`。客户端 MUST 解析 `auth_ok`、`session_sync`、`thinking_delta`、`answer_delta`、`answer_done`、`error` 帧。

#### Scenario: 连接 gateway-app 而非 voice-service

- **WHEN** App 建立胖宝诊疗 WebSocket
- **THEN** 连接目标主机 MUST 与 `apiBaseUrl` 一致（gateway-app-server）
- **AND** 路径 MUST 为 `/voice/clinic/ws`（或含 `apiBaseUrl` path 前缀的等价路径）

#### Scenario: 首帧 auth

- **WHEN** clinic WS 握手成功
- **THEN** 客户端 MUST 发送 `auth` 帧且 MUST NOT 在收到 `auth_ok` 前发送 `question`

#### Scenario: 未登录不发 question

- **WHEN** 用户未登录（无有效 accessToken）
- **THEN** 客户端 MUST NOT 建立可提问的 clinic WS 连接；若服务端返回 40301 MUST 引导登录

#### Scenario: 后台断开

- **WHEN** App 进入后台且 clinic WS 已连接
- **THEN** 客户端 MUST 主动断开 WebSocket

#### Scenario: 流式展示回答

- **WHEN** 收到 `answer_delta` 序列后以 `answer_done` 结束
- **THEN** UI SHALL 逐字/逐段更新回答区域

#### Scenario: session_sync 恢复历史

- **WHEN** 收到 `session_sync` 且 `turns` 非空
- **THEN** App SHALL 将已完成轮次填充至聊天 `_items`（user 问 + assistant 答 + 免责声明）
- **AND** MUST NOT 为历史轮次渲染 thinking（服务端未提供）

### Requirement: Flutter SHALL 实现 thinking 展示交互规范

胖宝诊疗页 thinking 区域 MUST：默认最多可见 **5 行**，不足 5 行时弹性高度；流式过程中 MUST 自动滚动至**最新** thinking 行（折叠视口 **底对齐** 最新内容，对齐 `home_voice_message_strip` 内层 `jumpTo(maxScrollExtent)` 模式）；超过 5 行 MUST 折叠，用户点击可展开全部或在折叠区内局部滚动；折叠态 MUST 使用内层 `ScrollController` 且 **MUST NOT** 使用 `NeverScrollableScrollPhysics` 阻止底对齐滚动；用户手动上滑固定（pin scroll）后 MUST 停止自动滚动直至用户回到底部或点击「跟随最新」。

#### Scenario: 流式自动滚动至最新行

- **WHEN** thinking 流式追加且用户未 pin scroll
- **THEN** 折叠视口 SHALL 展示最新 thinking 行（非顶部旧行）
- **AND** 内层 scroll offset SHALL 跳转至 `maxScrollExtent`

#### Scenario: 用户 pin 后停止跟随

- **WHEN** 用户上滑 thinking 区域内层 scroll 以查看较早内容
- **THEN** 后续 `thinking_delta` MUST NOT 强制跳回底部，直至用户点击「跟随最新」或滚回底部

#### Scenario: 跟随最新恢复

- **WHEN** 用户已 pin 且点击「跟随最新」
- **THEN** thinking 区域 SHALL 跳至最新内容并恢复流式 auto-scroll

### Requirement: Flutter SHALL 展示 clinic_ai 额度

`ai_quota_models` MUST 扩展 `clinicAi` 字段。诊疗页额度 hint MUST 展示 **「本月胖宝诊疗剩余 N 次」**（或等价 `AiQuotaRemainingHintFeature.clinicAi` 文案）。收到 WS `error` code **40302** 或 HTTP 40302 MUST 弹框 **「本月额度已用完」**。code **40301** MUST 引导登录。

#### Scenario: 胖宝诊疗额度展示

- **WHEN** 用户进入胖宝诊疗页且额度 API 返回 clinicAi remaining=5
- **THEN** UI SHALL 展示「本月胖宝诊疗剩余 5 次」

#### Scenario: 胖宝额度用尽

- **WHEN** clinic WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」
