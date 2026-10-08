## Context

设备详情页 `resource/public/history.html`（`/device/admin/history/{deviceNo}`）已具备：

- Hub Admin JWT（进页）；
- AI 调试台：`POST /device/app/api/username_login` 取得 App JWT，与 Admin 分钥；登录后校验账号 `deviceNo` 与页内设备是否一致。

小智绑定 App API（`xiaozhi-mcp-multi-device-binding`）已提供列表/添加/改备注/删除；gateway 已反代 `/device/app/api/xiaozhi-mcp/*`。本变更只补 Hub UI 门禁与调用。

## Goals / Non-Goals

**Goals:**

- 详情页「小智 MCP 绑定」卡片：App 登录后才能增删改查。
- 调用既有 App API；展示脱敏 token；添加字段为 `mcpToken` + `alias`。
- 登录账号 `deviceNo` ≠ 当前页时：告警且禁止添加（比调试台「仅 alert」更严，避免串号写绑定）。

**Non-Goals:**

- 新增 Admin API 或按 deviceNo 跨 wx 全量列表。
- Flutter；改绑定表 / Manager；完整 token 回显。
- 新 App 路由（无 usage 确认项）。

## Decisions

### D1. 复用 App 登录 + App API

- **选择**：与调试台共用或并列 App 登录态（同一 `sessionStorage` App token 键更佳，避免重复登录）；绑定请求 `Authorization: Bearer <App JWT>`。
- **理由**：`wx_id` / `deviceNo` 由服务端从会话与 wx 绑机推导，无需 Admin 侧猜归属。
- **备选**：Admin API 按 deviceNo → 本变更明确不做。

### D2. 设备号一致性：禁止添加

- **选择**：App 登录响应中的 `deviceNo`（或随后 detail）与 URL `deviceNo` 比较；不一致则 UI 禁用「添加」，并提示串号；列表/删除若服务端按 wx 过滤仍可展示该 wx 自己的绑定（实现可仅禁用写操作）。
- **理由**：绑定落点由服务端取「当前 wx 已绑宝宝」，串号会导致音箱挂到错误宝宝。

### D3. Token 仅脱敏

- **选择**：UI 只渲染接口返回的 `tokenMask`；表单添加时输入明文 token，提交后不再回显全文。
- **理由**：与 App API 契约一致，降低 Hub 截图泄露风险。

### D4. UI 位置

- **选择**：`history.html` 独立卡片（可置于 AI 调试台附近），表单：备注 + token + 添加；表格：alias / tokenMask / 改备注 / 删除。
- **理由**：详情页已有分区卡片模式，改动面可控。

### D5. Usage / 网关

- 无新 App/Admin 路由；反代已覆盖。tasks 仅自检反代前缀与未误改 `maintenance_skip`。

## Risks / Trade-offs

- [运维需持有用户密码] → 接受；与调试台相同运维模型。
- [仅见当前登录 wx 的绑定] → 文档/UI 注明；跨账号全设备视图需另案 Admin API。
- [App token 过期] → 提示重新登录；与调试台一致。

## Migration Plan

1. 部署含更新后的 `history.html` 的 gateway-app 静态资源即可。
2. 回滚：回退静态页；后端无强制迁移。

## Open Questions

- App token 与调试台是否强制共用同一 storage key：实现时优先共用，减少二次登录。
