## Context

已有 `xiaozhi_mcp_binding`（wx_id / device_no / mcp_token / alias）与 App CRUD、Hub App 登录绑定 UI、token normalize。token 会变，不能表达「同一音箱」；产品要求一宝宝多音箱、一音箱一宝宝，并以 MAC 为稳定 ID；同 MAC 再添加 = 更新 token。

Manager 仅有 `sessions[token]`，**有会话 ≠ 已连通**（Bridge 可能在退避重连）。Hub 需要绿/红灯展示真连通。

喂养 `need_confirm` 续聊依赖小智再次 `tools/call`；短确认/取消词若不在工具描述中，智能体常自行应答，voice `pendingConfirm` 收不到第二轮。

## Goals / Non-Goals

**Goals:**

- 添加必填 `speakerMac`；规范化为小写冒号分隔六段（`3c:dc:75:fc:7f:c4`）。
- `speaker_mac` 全局唯一；同一 `device_no` 允许多 MAC。
- 同 MAC + 当前宝宝：更新 token（normalize 后）、更新 alias（若请求提供）、mcp 侧停旧桥启新桥。
- 同 MAC + 其他宝宝：拒绝。
- Hub 表单/列表支持 MAC；App API 列表返回 `speakerMac`。
- mcp 进程内缓存 `token → connected`；内部 HTTP 批量查询；device 列表返回 `connected`；Hub 绿/红灯。
- `baby_feeding_advisor` 工具描述覆盖喂养确认/取消短句，促使多轮续聊仍调用本工具。

**Non-Goals:**

- 限制一宝宝只能一台音箱。
- 用 MAC 参与小智拨号 URL（拨号仍只用 token）。
- Flutter 工程内实现（仅后端+Hub；契约供 Flutter 后续对接）。
- Redis / cachekit 存连接态（本 change **明确不上 Redis**）。
- 第三态「connecting / 黄灯」。
- 新增独立 confirm MCP 工具；改动 voice/Python 澄清协议或 `pendingConfirm` TTL。

## Decisions

### D1. MAC 规范化

- 去空白、转小写；允许输入 `3C-DC-75-FC-7F-C4` / `3cdc75fc7fc4` 等，统一输出 `3c:dc:75:fc:7f:c4`。
- 非法格式 → `speakerMac 无效`。

### D2. 同 MAC 再添加 = 更新 token

- **选择**：`Add` 按规范化 MAC 查行：无则 insert；有且 `device_no` 等于当前 wx 宝宝则 update token/alias；有且 device 不同则错误「该音箱已绑定其他宝宝」。
- **理由**：换密无需单独 Update API；与产品「再添加即更新」一致。
- mcp 通知：若 token 变化，先 `BindingRemove(旧)` 再 `BindingUpsert(新)`；token 未变可只更新 alias 不惊动桥（或仍 Upsert 幂等）。

### D3. token 全局唯一仍保留

- 规范化后的 `mcp_token` 仍 `uk` / 业务去重；更新时排除自身 id 再查冲突。

### D4. 历史无 MAC 行

- Ensure 加列时允许临时空串；加唯一索引前：**删除或要求运维清空**无 MAC 的测试行；新写入 MAC 必填，禁止再写空 MAC。
- 实现：`EnsureXiaozhiMcpSpeakerMacColumn` + 对空 MAC 行 `DELETE`（或迁移失败 fail-fast 打日志），再 `UNIQUE(speaker_mac)`。

### D5. Hub（MAC + 连接灯）

- 输入框：MAC + token + 备注；列表列：MAC、alias、tokenMask、**连接状态灯**。
- `connected === true` → 绿灯；否则 → 红灯（含 false、字段缺失、查询失败由服务端已置 false）。

### D6. 连接状态：进程内存 + 内部 HTTP（方案 A）

```
Bridge dial 成功进入 readLoop → Manager 标 token connected=true
断线 / Remove / 停桥              → connected=false 或删条目

device List bindings
  → 收集规范化 tokens
  → POST mcp /xiaozhi-mcp/internal/api/bindings/connection-status { tokens: [] }
  → 每条绑定填 connected bool
```

- **选择**：状态只活在 xiaozhi-mcp-service 内存；device **不**持久化、**不**读 Redis。
- **理由**：已约定 `replicas=1`；与现有 internal secret HTTP 一致；避免新 Redis 键与负责人确认。
- **语义**：`connected=true` **仅当** 当前与小智 MCP WebSocket 处于读循环中；会话存在但拨号失败/重连退避 → `false`。
- **批量接口**：请求 `tokens[]`（或可选按 binding id）；响应 `map[token]bool` 或 `[{mcpToken, connected}]`；未知 token → `false`。
- **鉴权**：同 upsert/remove，`DEVICE_GATEWAY_INTERNAL_SECRET`。
- **device 失败语义**：mcp 不可达或超时 → 列表仍成功，相关项 `connected=false`（打 WARN，不 5xx），避免假绿。
- **clients**：`internal/clients/xiaozhimcp` 新增 `ConnectionStatus(ctx, tokens []string) (map[string]bool, error)`；device 不得 import `mcpbridge`。
- **键**：规范化后的 `mcp_token`（与拨号/session key 一致）；MAC 不参与连接态。
- **token 轮换**：Remove 旧 token 状态；新 token 从 false 直至 dial 成功。
- **进程重启**：内存清空 → 短暂全红，reconcile/dial 成功后变绿——可接受。

### D7. App 列表字段

- 列表项增加 `speakerMac`、`connected`（bool）。App 与 Hub 共用同一 List API；Flutter 可后续消费，本 change 不改 Flutter。

### D8. 喂养确认/取消：仅扩工具描述

- **选择**：在 `internal/services/mcpbridge/tools.go` 的 `chatToolDescription` 增加选路条件（确认 + 取消示例词），**不**新增工具名、**不**改 `ChatHandler` / `ChatViaVoiceWS`。
- **理由**：断点在小智是否 `tools/call`；voice 侧 `pendingConfirm` + `conversation_id` 已具备续聊能力。
- **文案要点**（实现时可润色，语义须保留）：
  - 用户对上一条由本工具发起的喂养相关操作做确认或取消时，必须调用本工具；
  - 确认词：是的、对的、确定、好的、可以、没问题等；
  - 取消词：取消、不要、算了、不对等；
  - 将用户原话放入 `transcript`，不要自行回答。
- **生效**：部署 mcp 后需小智重新 `tools/list`（Bridge 重连或平台刷新 MCP）。
- **误召回**：闲聊短句也可能进工具；无 pending 时由既有 Python/voice 路径消化，可接受。

## Risks / Trade-offs

- [BREAKING 旧客户端无 MAC] → App/Hub 同步改；文档说明。
- [脏数据空 MAC] → 启动清理策略见 D4。
- [更新抢绑] → 其他宝宝占用 MAC 明确拒绝，不静默覆盖。
- [mcp 重启全红] → 可接受；单副本假设破了则内存态不一致（非本 change 范围）。
- [列表多一次内部 HTTP] → 绑定数通常很小；超时短、失败降级为红灯。
- [描述无法 100% 保证智能体选工具] → 仅靠描述启发式；若线上仍漏召可再考虑独立小工具（非本 change）。

## Migration Plan

1. 部署 device：加列、清空空 MAC、唯一索引、Add upsert、列表拼 `connected`。
2. 部署 xiaozhi-mcp-service：连接标志 + status 内部 API + 工具描述更新；必要时重连 Bridge 以刷新 tools/list。
3. 部署 gateway-app 静态 Hub（MAC 表单 + 绿/红灯）。
4. 回滚：保留列可空会破坏唯一语义，宜前向修数据；去掉灯仅 UI/字段回退即可；工具描述可单独回滚文案。

## Open Questions

- 无（同 MAC 更新 alias：随本次请求覆盖；连接态方案 A + 仅绿/红；确认/取消仅扩描述：已定）。
