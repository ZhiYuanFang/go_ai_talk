## Why

`xiaozhi-mcp-permanent-unlock` 已提供 wx 维永久能力与 Add 门禁，但缺少「体验一次」；且仅卡 Add 时，试用/退款后已建 Bridge 仍可 `tools/call`，等于白嫖。需补齐限时试用，并在工具调用路径强制校验，过期懒停桥、保留绑行。

## What Changes

- 权益支持 **试用 24h**（`expires_at`）与 **永久**（`expires_at=0`）；每 wx **仅可领取试用一次**。
- **首次成功 Add** 时：若无有效权益且试用未用 → claim 24h 后再写绑定（失败则整笔 Add 失败）；已过期且未付费 → 拒绝 Add。
- **tools/call 硬闸门**：执行喂养工具前校验绑定所属 wx 仍有有效权益；无效则返回 tool 错误，**不**调用 voice。
- **懒停桥**：在上述拒答路径上尽力 `Remove` 当前 Bridge；绑行 **保留**（token/MAC 不删）；付费/再开通后可复建桥。
- Bridge / Upsert / 内部 list 携带 **wxId**，供 mcpbridge 经 `clients/cash` 查开通态（禁止直查 cash 库）。
- App `unlock` 响应补充 `trialAvailable` / `expiresAt`（或等价字段）。
- 退款/撤销导致未开通时，同样依赖 tools 闸门 + 懒停（与试用过期同语义）。

## Capabilities

### New Capabilities

- `xiaozhi-mcp-trial`：wx 一次试用 24h、首次 Add claim、开通态字段扩展。
- `xiaozhi-mcp-usage-gate`：tools/call 权益校验、懒停桥、绑定携带 wxId、过期保留绑行。

### Modified Capabilities

- （依赖进行中的 `xiaozhi-mcp-permanent-unlock` 权益表与 Add 门禁；本变更在其之上扩展，不单独改基线 `openspec/specs/` 中尚不存在的归档名。）

## Impact

- **cash-service**：权益 schema（`expires_at`、试用记账）、claim API/内部能力、`unlock` 字段、Add 侧可调用的 claim。
- **device-service**：Add 流程与 cash 协作（先 claim 或校验）；Upsert/内部 list 传 `wxId`。
- **xiaozhi-mcp-service / mcpbridge**：会话持有 wxId；tools/call 调 cash；拒答后懒停桥。
- **clients/cash**：供 mcpbridge（及既有 device）查询开通态。
- **Flutter / Hub**：展示试用可用与到期；行为依赖服务端闸门。
