## Why

小智 MCP `token` 会轮换，不能作为音箱稳定身份；业务需要「一个宝宝可绑多台音箱，但一台音箱只能属于一个宝宝」。因此以音箱 **MAC**（如 `3c:dc:75:fc:7f:c4`）为全局唯一标识；同一 MAC 再次添加时 **自动更新 token**（及可选备注），实现换密而不换绑。

运维还需在 Hub 一眼看到桥是否真连上小智：列表旁绿/红灯，避免只靠日志猜连通状态。

喂养澄清多轮中，用户仅说「是的」「对的」「确定」等短确认/取消词时，小智智能体常不调用 MCP 工具，导致 voice 侧 pending 续聊断掉；须在工具描述中显式要求此类短句仍调用 `baby_feeding_advisor`。

## What Changes

- **BREAKING（App 添加）**：`POST /device/app/api/xiaozhi-mcp/bindings` 增加必填 `speakerMac`；服务端规范化并校验 MAC 格式。
- `xiaozhi_mcp_binding` 表增加 `speaker_mac` 列与 **全局唯一索引**；一个宝宝（`device_no`）可有多行（多 MAC）。
- 添加语义：**同 MAC 已存在** → 若归属当前宝宝则更新 `mcp_token`（及本次提交的 alias，若有），并通知 mcp Remove(旧 token)+Upsert(新)；若归属其他宝宝 → 拒绝。
- 新 MAC → 插入新行（仍 normalize token、token 全局去重）。
- Hub 详情页小智表单增加 MAC 输入；列表展示 MAC。
- 内部全量 list / Upsert 载荷可带 MAC（观测用）；拨号仍仅依赖 token。
- **连接状态（本 change 一并交付）**：
  - xiaozhi-mcp-service 进程内按 **规范化 token** 维护 `connected`（内存临时缓存；**不上 Redis**）。
  - 新增内部 HTTP 批量查询连接状态；device-service 在绑定列表中按 token 填充布尔字段。
  - Hub 列表：`connected=true` 绿灯，否则红灯（含重连中、无会话、mcp 不可达）。
  - 仅绿/红两态，不引入「黄灯 / connecting」。
- **喂养确认/取消工具描述（本 change 一并交付）**：
  - 扩展 `baby_feeding_advisor` 的 `chatToolDescription`：用户对上一条喂养相关操作确认或取消时必须调用本工具。
  - 确认词示例：是的、对的、确定、好的、可以、没问题；取消词示例：取消、不要、算了、不对。
  - 不新增独立工具；Handler / voice WS 路径不变，仍将用户原话作 `transcript` 传入。

## Capabilities

### New Capabilities

- `xiaozhi-mcp-speaker-mac`：音箱 MAC 规范化与全局唯一、一宝宝多音箱、同 MAC 再添加自动更新 token、App/Hub 入参与列表展示；token 级连接状态（mcp 内存 + 内部 HTTP）与 Hub 绿/红灯；以及喂养确认/取消短句的 MCP 工具选路描述。

### Modified Capabilities

- （基线未归档 `xiaozhi-mcp-binding`；行为增量以本 change 新 capability 为准，实现时改同一套 App API/表。）

## Impact

- **代码**：`schema`/`entity`/`dao`、`AddXiaozhiMcpBinding`（upsert by MAC）、App API 契约、Hub `history.html`、写路径 mcp 通知（token 变更时停旧桥）；`mcpbridge` Manager/Bridge 连接标志、内部 status API、`clients/xiaozhimcp` 批量查询、device 列表拼装 `connected`；`mcpbridge/tools.go` 工具描述文案。
- **数据**：已有无 MAC 行需迁移策略（空 MAC 禁止新唯一约束冲突：清理或一次性补录后再加索引）。
- **客户端**：Flutter/Hub 添加必须带 MAC；旧客户端只传 token 将失败（BREAKING）；列表新增 `connected`（及 `speakerMac`）。
- **usage**：无新 App 路由 path（同 path 增字段）；内部 status 接口不经 gateway-app；不改 `maintenance_skip`。
- **部署**：依赖 `replicas=1` 的 xiaozhi-mcp-service（内存态）；mcp 重启后短暂全红直至重新 dial。
