## Why

Hub「小智 MCP 开通」页首版仅有 SKU 配置与按 wxId 手工授/撤销，运维无法浏览已开通账号，只能凭记忆填号操作。开通功能管理已有「开通快照」（昵称 + 行内撤销），小智 MCP 应对齐该运维体验。

## What Changes

- 新增 Admin **开通人员列表**接口：分页读取 `xiaozhi_mcp_entitlement` 快照（含有效/已失效），经 ucg 批量补昵称。
- 静态页 `cash-xiaozhi-mcp-admin.html` 增加 **区 C · 已开通**：表格展示 wxId、昵称、开通方式、状态、到期、开通/更新时间、channel_ref；支持刷新与分页。
- 行内 **撤销**：仅当该行仍有效且 `unlock_method=admin` 时展示；复用既有 `POST .../grants/revoke`，撤销后刷新列表。
- 区 B 手工授/撤销成功后 SHOULD 刷新区 C，避免运维看不到刚授的人。

## Capabilities

### New Capabilities

- `xiaozhi-mcp-unlock-admin-list`：Hub 小智 MCP 开通页的已开通人员快照列表（含昵称）与行内手工授撤销。

### Modified Capabilities

- （无。基线 `openspec/specs/` 尚无已归档的 `xiaozhi-mcp-unlock-admin`；本变更为独立增量能力，不改支付/App 开通语义。）

## Impact

- **cash-service**：新增 Admin list 领域函数与控制器；读 `xiaozhi_mcp_entitlement`；经 `clients/ucg` 拉昵称（与 feature activations 同模式）。
- **api/v1**：新增 Admin GET（如 `/cash/admin/api/xiaozhi-mcp/entitlements`），`g.Meta` 完整。
- **Hub 静态页**：`resource/public/cash-xiaozhi-mcp-admin.html` 区 C。
- **gateway-app**：路径落在既有 cash Admin 反代与 Admin 口令鉴权内；**非** App Bearer 接口，不涉及 usage 统计策略变更。
- **不改**：SKU 读写、支付履约、App unlock/建单、device Add 门禁、撤销业务规则（仍仅 admin 手工授可撤）。
