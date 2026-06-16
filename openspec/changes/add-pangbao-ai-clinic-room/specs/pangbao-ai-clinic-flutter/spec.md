## ADDED Requirements

### Requirement: Flutter SHALL 提供胖宝 AI 诊室入口与页面

App（`flutter_ai_talk`）MUST 在首页 `home_immersive_header.dart` **新增** **胖宝** 入口（`Icons.pets`，tooltip「胖宝」），跳转至新页面 `pangbao_ai_screen.dart`。页面 MUST 支持文本输入提问并通过 WebSocket 展示流式 thinking 与 answer。首页 **MUST 保留** 原 **趋势** 入口（`Icons.insights`，tooltip「趋势」→ `/trends`），胖宝入口 **MUST NOT** 替换或隐藏趋势入口。

#### Scenario: 从首页进入胖宝

- **WHEN** 用户在首页点击胖宝入口
- **THEN** App SHALL 导航至胖宝 AI 页面

#### Scenario: 从首页进入趋势

- **WHEN** 用户在首页点击趋势入口
- **THEN** App SHALL 导航至趋势图表页面（`/trends`）

### Requirement: Flutter SHALL 使用 ClinicWsClient 经 gateway-app 连接 WebSocket

App MUST 实现 `clinic_ws_client.dart`。连接 URL MUST 使用 `wsClinicUrl` / `wsClinicUrlEffective`，默认由 `apiBaseUrl`（gateway-app-server 主机）推导为 `wss://{host}/voice/clinic/ws`（对齐 `wsVoiceAsrUrlEffective` 模式）。客户端 **MUST NOT** 配置或连接 voice-service 内网地址。连接成功后 MUST 先发送首帧 `type=auth`（`accessToken` + `deviceNo`，与 history WS / UCG chat 一致），收到 `auth_ok` 后方可发送 `question`。WS 生命周期 MUST 对齐 UCG chat：App 进入后台 MUST disconnect；回前台 MAY 重连且 MUST 重新 `auth`。客户端 MUST 解析 `auth_ok`、`thinking_delta`、`answer_delta`、`answer_done`、`error` 帧。

#### Scenario: 连接 gateway-app 而非 voice-service

- **WHEN** App 建立胖宝 WebSocket
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

### Requirement: Flutter SHALL 实现 thinking 展示交互规范

胖宝页 thinking 区域 MUST：默认最多可见 **5 行**，不足 5 行时弹性高度；流式过程中 MUST 自动滚动至最新 thinking 行；超过 5 行 MUST 折叠，用户点击可展开全部或在折叠区内局部滚动；用户手动上滑固定（pin scroll）后 MUST 停止自动滚动直至用户回到底部或点击「跟随最新」。

#### Scenario: 流式自动滚动

- **WHEN** thinking 流式追加且用户未 pin scroll
- **THEN** thinking 区域 SHALL 自动滚至最新内容

#### Scenario: 用户 pin 后停止跟随

- **WHEN** 用户上滑 thinking 区域以查看较早内容
- **THEN** 后续 `thinking_delta` MUST NOT 强制跳回底部，直至用户明确恢复跟随

### Requirement: Flutter SHALL 展示免责声明

每条 AI 回答（`answer_done` 后）UI MUST 展示文案：**「本回答仅供参考，不能替代医生诊断」**。

#### Scenario: 回答完成后展示免责

- **WHEN** 一次问答流式完成
- **THEN** 该条回答下方 SHALL 显示上述免责声明

### Requirement: Flutter SHALL 使用独立胖宝 consent

胖宝功能 MUST 使用独立 consent 键 `pangbao_ai_consent_v1`，与首页喂养 AI consent 分离。首次进入胖宝页且未 consent 时 MUST 展示同意流程；未同意 MUST NOT 发送 `question`。

#### Scenario: 首次进入需 consent

- **WHEN** 用户首次打开胖宝页且本地无 `pangbao_ai_consent_v1`
- **THEN** App SHALL 展示 consent 对话框且 MUST NOT 发送 `auth`/`question` 直至用户同意

### Requirement: Flutter SHALL 展示 clinic_ai 额度

`ai_quota_models` MUST 扩展 `clinicAi` 字段。收到 WS `error` code **40302** 或 HTTP 40302 MUST 弹框 **「本月额度已用完」**。code **40301** MUST 引导登录。

#### Scenario: 胖宝额度用尽

- **WHEN** clinic WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」
