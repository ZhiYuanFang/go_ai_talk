## 1. 详情页 UI 与 App 登录门禁

- [x] 1.1 在 `history.html` 增加「小智 MCP 绑定」卡片（备注+token 表单、列表区、状态提示）
- [x] 1.2 复用或对接页内 App 登录态（优先与 AI 调试台共用 App token storage）；未登录禁用写操作并提示
- [x] 1.3 登录后比对账号 deviceNo 与页内 deviceNo：不一致告警并禁止添加

## 2. 调用既有 App API

- [x] 2.1 列表：App Bearer `GET /device/app/api/xiaozhi-mcp/bindings`，展示 alias + tokenMask（不渲染全文 token）
- [x] 2.2 添加：`POST` mcpToken+alias；成功刷新列表
- [x] 2.3 改备注：`PUT .../bindings/{id}/alias`；删除：`DELETE .../bindings/{id}`

## 3. 自检

- [x] 3.1 确认未新增 App/Admin 业务路由；gateway 已反代 `/device/app/api/xiaozhi-mcp/*`；未改 `maintenance_skip.go`
- [x] 3.2 手工：未登录不可绑；同设备登录可增删改查；串号登录不可添加
