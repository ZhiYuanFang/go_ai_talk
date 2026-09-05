## ADDED Requirements

### Requirement: 服务端不得再暴露小贴士 SSE

voice-service 与 gateway-app MUST NOT 再对外提供 `POST /device/tip/generate`（及 `/device/tip/*` 反代）。`TipCtrl`（或等价）MUST NOT 再注册；`VoiceService.TipStream` 与仅服务于 tip 的 Python 客户端方法 MUST 删除或不可达。

#### Scenario: tip generate 不可达

- **WHEN** 客户端向 gateway 发起 `POST /device/tip/generate`
- **THEN** 请求 MUST NOT 由 tip SSE 控制器处理（典型为 404 或未绑定路由），且 voice-service HTTP Bind 列表 MUST NOT 包含 `NewTipCtrl()`（或等价）

#### Scenario: TipStream 从运行时移除

- **WHEN** 检索 voice 服务实现与 contracts 中的 tip 流式入口
- **THEN** MUST NOT 存在可被 HTTP 控制器调用的 `TipStream` 业务路径

### Requirement: 运维页不得再调用 tip generate

`resource/public/history.html`（及本仓内其它运维静态页若存在同等调用）MUST NOT 再请求 `/device/tip/generate`，相关调试 UI MUST 删除或改为不可用状态。

#### Scenario: history 运维无 tip 请求

- **WHEN** 审查 `resource/public/history.html` 源码
- **THEN** MUST NOT 出现对 `/device/tip/generate` 的 `fetch`/XHR 调用

## REMOVED Requirements

### Requirement: tip 请求路径为 /device/tip/generate（基线 v3.0.0）

**Reason**：产品已用本地预测/care-alert 替代 SSE 小贴士；Flutter 与运维均不再需要该接口。  
**Migration**：客户端改用本地预测条/care-alert；运维删除 tip 调试入口。无替代 SSE 路径。

### Requirement: 系统 SHALL 暴露 POST /device/tip/generate SSE（基线 v3.0.0 等价）

**Reason**：同上。  
**Migration**：删除服务端实现与 gateway 反代；无兼容桩。
