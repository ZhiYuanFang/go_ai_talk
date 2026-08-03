## ADDED Requirements

### Requirement: voice admin SHALL list per-wx effective quota with identity fields

voice-service MUST 提供 `GET /voice/admin/api/ai-quota/users`，认证 MUST 为 Header `X-Admin-Password` 等于 `voice.admin.password`（gateway 经 `VOICE_ADMIN_PASSWORD` 注入）。

Query MUST 支持：`page`（从 1）、`pageSize`（默认 20，最大 100）、**`deviceNo`**（可选，按设备号过滤）。实现 MUST 经 **`DeviceAdmin().ListWxPage`**（或等价 device 契约）取得真实 wx 分页全集（排除模拟号），MUST NOT 在 voice 进程直查 device/history 库表。

响应 MUST 为分页结构 `{ list, total, page, pageSize }`。`list[]` 每项 MUST 含：

- `deviceNo`（string）
- `wxId`（int64，对应 wx.id）
- `account`（string，对应 wx.account）
- `babyName`（string，无档案时为空串）
- `voiceAi`：`{ used, limit }`（当月上海时区桶；`limit` 为有效上限 = 正数 override ∨ 全局默认）
- `clinicAi`：`{ used, limit }`（同上）

`used` MUST 来自既有 Redis usage 键（经 `cachekit`）；键不存在时 MUST 为 0。本接口 MUST NOT 修改用量或 override。本接口 MUST NOT 返回 polish 字段。

改上限 MUST 继续使用既有 `PUT /voice/admin/api/ai-quota/user`；当提交的某 feature 上限等于当前全局默认时，该 feature MUST 清除 override（写空/NULL），使该用户后续跟随全局默认。

#### Scenario: 分页返回身份与双额度

- **WHEN** 管理员请求 `/voice/admin/api/ai-quota/users?page=1&pageSize=20` 且 device 域存在真实 wx
- **THEN** 响应 `list` 每项 SHALL 含 deviceNo、wxId、account、babyName 以及 voiceAi/clinicAi 的 used 与 limit

#### Scenario: 按 deviceNo 过滤

- **WHEN** 管理员请求 `deviceNo` 等于某已绑定设备号
- **THEN** 返回列表 SHALL 仅包含该设备号匹配的 wx 行（在契约过滤语义下）

#### Scenario: 无 override 时 limit 为全局默认

- **WHEN** 某 wx 无 clinic override 且全局 clinicAiMonthlyLimit=30
- **THEN** 该行 `clinicAi.limit` SHALL 为 30

#### Scenario: 有 override 时 limit 为覆盖值

- **WHEN** wxId=1001 的 voice override 为 10 且当月 voice 已用 3
- **THEN** 该行 `voiceAi.limit` SHALL 为 10 且 `voiceAi.used` SHALL 为 3

#### Scenario: 口令错误拒绝

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 返回用户额度列表

#### Scenario: 保存等于默认则清除 override

- **WHEN** 管理员 PUT user 将某 wx 的 clinicAiMonthlyLimit 设为与当前全局 clinic 默认相同的正整数
- **THEN** 该 wx 的 clinic override SHALL 被清除（后续 limit 跟随全局默认）

## MODIFIED Requirements

### Requirement: voice admin SHALL configure global default and per-user override locally

voice-service MUST 提供 `GET/PUT /voice/admin/api/ai-quota/default` 与 `GET/PUT /voice/admin/api/ai-quota/user`（query/body 含 `wxId`），以及 **`GET /voice/admin/api/ai-quota/users`**（用户额度分页列表，见 ADDED 要求）。认证 MUST 为 Header `X-Admin-Password` 等于 `voice.admin.password`（gateway 经 `VOICE_ADMIN_PASSWORD` 注入）。voice-service MUST 本地持久化至 `ai_voice_voice`，MUST NOT 转发 device 或 ucg 写配额。PUT default MUST 接受 `voiceAiMonthlyLimit` 与 `clinicAiMonthlyLimit`（正整数）。PUT user MUST 接受 optional 两字段；空值 SHALL 表示清除该 feature override。运维浏览全量用户额度 MUST 使用 users 列表 API，而非依赖仅按 wxId 点查的 UI 作为唯一入口。

#### Scenario: 管理员修改全局胖宝默认

- **WHEN** 管理员 PUT default 为 voiceAi=5、clinicAi=30
- **THEN** voice 权威配置 SHALL 更新且新用户 check clinic_ai SHALL 使用 limit=30

#### Scenario: voice admin 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改配置

#### Scenario: 列表 API 与单人写并存

- **WHEN** 管理员先 GET users 再对其中一行 PUT user 修改 voiceAiMonthlyLimit
- **THEN** 后续 GET users 同 wx 行的 `voiceAi.limit` SHALL 反映新有效上限
