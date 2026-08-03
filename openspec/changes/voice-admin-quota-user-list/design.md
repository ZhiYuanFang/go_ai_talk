## Context

当前 `voice-admin.html` 仅有全局默认 + 手输 wxId 的单人 override。额度权威在 voice（`ai_quota_default` / `ai_quota_user_override` + Redis `ai:usage:{feature}:{wxId}:{YYYYMM}`）；wx 身份在 device（`ListWxPage`，已排除模拟号）；`baby_name` 在 device `user` 画像。运维需要按设备号浏览并改上限，单人表单与列表改额度功能重叠。

约束：跨域 MUST 走契约（voice → `DeviceAdmin()` HTTP）；Redis 用量读 MUST 经 `cachekit`；Admin 认证沿用 Hub JWT + gateway 注入 `X-Admin-Password`；不新增 App 对外 HTTP；不新增背景 ticker。

## Goals / Non-Goals

**Goals:**

- 分页展示 device 全部真实 wx 的喂养/胖宝当月已用与有效上限。
- 表格第一列为 `deviceNo`；一行含两额度（已用/上限分列）；支持按 `deviceNo` 查询；行内改上限。
- 去掉单人 override 模块；保留全局默认。
- 说明中列明喂养/胖宝受影响业务接口。

**Non-Goals:**

- 不改 App/internal check·consume·degraded 语义与 Redis 键。
- 不提供「改已用」或跨月用量编辑。
- 不改造 UCG 润笔 admin。
- 不为列表新建独立 Redis 缓存层（页内 MGET/逐条 GET 现有 usage 键即可）。

## Decisions

### 1. 列表 API 归属 voice admin

- **选择**：新增 `GET /voice/admin/api/ai-quota/users?page&pageSize&deviceNo=`（可选弱 `q`），由 voice-service 聚合后返回。
- **理由**：额度 used/limit 权威在 voice；身份经已有 `DeviceAdmin().ListWxPage` 拉取，符合域边界。
- **备选**：前端分别调 `/device/admin/api/wx/list` + 多次 `/voice/admin/api/ai-quota/user` —— N+1 且无 used，运维体验差。

### 2. 用户全集与过滤

- **选择**：全集 = `ListWxPage`（`is_simulated=0`）。`deviceNo` 为专用查询参数，对 `device_no` 做前缀/模糊匹配（实现可复用或收窄现有 `q` 中的 deviceNo 分支）。可选 `q` 继续覆盖 wxId/account（弱搜索，非必须展示复杂搜索框）。
- **理由**：与现有 admin wx 列表一致；运维主路径是按设备号找人。
- **备选**：仅有 override / 本月有用量 —— 与「全部用户」产品要求不符。

### 3. 行字段与列顺序

表格列（左→右）：

1. `deviceNo`
2. 喂养已用（只读）
3. 喂养上限（可改）
4. 胖宝已用（只读）
5. 胖宝上限（可改）
6. `wxId`（`wx.id`）
7. `wx.account`
8. `babyName`

- **有效上限**：该 feature 存在正数 override 则用 override，否则全局默认。
- **已用**：当月上海时区桶 Redis；键不存在视为 0。
- **babyName**：优先扩展 device `ListWxPage` / `AdminWxListItem` 联查 `user.baby_name`（按 `device_no`）；无绑定设备或无档案时为空串。禁止 voice 直查 user 表。

### 4. 改上限写路径

- **选择**：行内保存调用既有 `PUT /voice/admin/api/ai-quota/user`（一次提交该行两个 feature 上限）。
- **等于全局默认时**：对该 feature **清除 override**（写 NULL），使后续改全局默认可惠及该用户。
- **理由**：复用写语义与口令校验；避免第二套写 API。
- **备选**：仅当显式点「恢复默认」才清除 —— 运维更易残留无意义 override。

### 5. UI 结构

- **保留**全局默认卡片。
- **删除**「单人 override（wxId）」整块（输入框/加载/清除 checkbox）。
- **新增**额度表 + 分页 + `deviceNo` 查询框；LLM 仍仅链至 `ai-model-admin.html`（不恢复页内 LLM Tab）。
- 页内简短说明列出喂养/胖宝影响接口（与 proposal Impact 一致），便于运维理解改上限后果。

### 6. 分页与性能

- page/pageSize 对齐现有 admin（默认 20，最大 100）。
- 对本页 wxId 批量解析 limit（一次读 default + 批量查 override 表）与 used（cachekit 对既有 usage 键 GET/MGET）；不引入新键命名。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| wx 量大时每页多次 Redis GET 延迟 | 页大小封顶 100；同页批量；可观测日志 |
| `ListWxPage` 无 babyName 时 N+1 画像 | **优先**扩展 device 列表项一次返回 babyName |
| 清 override 后「看起来」上限未变但语义变了 | 可选行标「个性化」；默认不强制，文档说明 |
| 基线 `voice-admin-ui` 仍写 LLM Tab | 本变更 MODIFIED 额度需求时与现行「链至 ai-model-admin」对齐，不恢复页内 LLM 编辑 |

## Migration Plan

1. 先上列表 API（只读）+ device 列表项 babyName（若需），再改 HTML。
2. 去掉单人模块与上线列表同发，避免双入口并存过久。
3. 回滚：恢复旧 HTML + 下线新 GET；override 表数据兼容，无需 migration。

## Open Questions

- （无阻塞）弱搜索 `q` 是否在 UI 暴露：默认仅 `deviceNo` 输入框；后端可保留 `q` 以复用 `ListWxPage`。
