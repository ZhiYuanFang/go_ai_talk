## Context

Python 已交付独立覆盖缺口收件箱（`agent_miss` JSONL + `/v1/admin/miss-inbox` list/delete），在终局失败回复前落原话与失败文案。与 `rule_gaps`（TIME/QTY 扩规则工单）及时间飞轮隔离。

Go Hub 已有规则缺口页范式：Hub HTML → `/voice/admin/api/rule-gaps*` → `PythonAIClient` → Python。意图智能体门户工单区目前仅规则缺口。设备 history 运维页路径为 `/device/admin/history/{deviceNo}`（`api-usage-stats` 等已用同深链）。

约束：Admin JWT + gateway 注入 `X-Admin-Password`；不直连 JSONL；Admin 不碰 usage；`/voice/admin/api/*` 反代已覆盖；不新增 Redis / 背景循环；不新增测试文件。

## Goals / Non-Goals

**Goals:**

- Hub 可分页查看覆盖缺口（原话、失败回复、fail_kind、设备、时间等）。
- 支持按 `status`（默认 open）与 `fail_kind` 筛选；单条物理删除。
- 非空 `device_no` 链到既有 history 运维页。
- 门户「规则工单」区可发现该入口。

**Non-Goals:**

- PATCH dismissed / bulk / 从 Hub 手工新增记录。
- 改 Python 落库判定或 fail_kind 枚举。
- 删除时自动灌飞轮或写 rule_gap。
- 新建 MySQL 表；合并入规则缺口页 Tab。

## Decisions

### D1：代理路径与 rule-gaps 对称，动作为 DELETE

| Hub | Python |
|-----|--------|
| `GET /voice/admin/api/miss-inbox` | `GET /v1/admin/miss-inbox` |
| `DELETE /voice/admin/api/miss-inbox/{id}` | `DELETE /v1/admin/miss-inbox/{miss_id}` |

查询透传：`offset`、`limit`、`status`、`fail_kind`（空串不过滤，与上游一致；默认列表传 `status=open`）。

**备选**：gateway 直打 Python → 否决（破坏 Hub→voice 口令惯例）。用 PATCH 标关闭 → 否决（上游仅物理删）。

### D2：静态页镜像 rule-gaps，非飞轮页

- 页：`miss-inbox-admin.html`；列：# / fail_kind / utterance / reply / device_no / conversation_id / status / created_at / 删除。
- 筛选：status（open / 全部）、fail_kind（空=全部；可预设已知 kind：`event_unresolved`、`other_unsupported`、`predict_uncertain`、`segment_unclear`）。
- 分页：同 rule-gaps 上一页/下一页。
- 文案：独立于规则缺口与时间飞轮；用途为发现答不出盲区；删=移出收件箱。

登记：`admin-modules.js`（`showInNav: false`）、`admin_static_pages.go`、`gateway_app_auth_exempt.go`。

### D3：device_no → history 深链

非空 `device_no` 渲染为链到 `/device/admin/history/{encodeURIComponent(deviceNo)}`（与 `api-usage-stats.html` 同构）；空则纯文本「—」。不新开 history API。

### D4：门户工单区第二卡

`intent-agent-admin.html` 规则工单区：规则缺口 + **覆盖缺口**（副文案可写「答不出 / 缺失语境」）。不进飞轮网格。

### D5：部署边界

须发：`voice-service`、`gateway-app`。依赖 python-ai-talk miss-inbox 已上线；未发时 list 失败 → adminFetch 可读错误。

## Risks / Trade-offs

- [Python 未发版] → 页内错误提示；门户其它卡不受影响
- [物理删除不可恢复] → 删除前 confirm；文案标明非 dismissed
- [fail_kind 日后新增] → 筛选保留自定义/空=全部，预设仅辅助
- [空 device_no] → 不链、显示「—」

## Migration Plan

1. 确认线上 python-ai-talk 已挂 `/v1/admin/miss-inbox`。
2. 发 voice-service + gateway-app。
3. 目视：门户工单两卡；列表/筛选/删；有 device_no 可进 history。
4. 回滚：撤静态与 voice 代理即可；不碰 Python JSONL。

## Open Questions

（无。卡片名「覆盖缺口」、默认 open、device_no 深链已拍板。）
