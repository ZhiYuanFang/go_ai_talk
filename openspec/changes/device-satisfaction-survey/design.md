## Context

满意度问卷要按「一次活动」收集用户觉得胖宝 App 做得满意的地方。选项由运维配置，可多选，也可只写意见。提交后该活动不再对该用户弹出。用户、宝宝昵称、设备号都在 device 域（`wx` / `user.baby_name`），适合放在 device-service，经 gateway-app 按子路径反代。现有 `/device/app/api/*` 不能做总前缀反代，需要像 `appointment` 一样显式登记 `/device/app/api/survey/*`。

负责人已确认两条 App 接口不计入功能使用统计。不为该读路径新增 Redis。

## Goals / Non-Goals

**Goals:**

- 活动级开关、上海时区起止时间、可编辑选项行。
- App 读取「是否展示」与提交；一次提交关闭该活动对该 wxId 的弹出。
- Hub 左编辑、右结果；结果含 wxId、宝宝昵称、设备号、选项快照、意见。
- 网关反代、登录鉴权、usage 精确排除。

**Non-Goals:**

- Flutter sheet（兄弟仓 `home-satisfaction-survey`）。
- 多题问卷、单选、评分量表。
- 问卷配置或结果的 Redis 读缓存。
- 修改既有 v1 接口字段。
- 用提交驱动推送或后台轮询。

## Decisions

### D1：三张 device 表，启动时幂等建表

- **选择**：沿用 `schema_appointment.go` 的 information_schema + `CREATE TABLE IF NOT EXISTS` 风格，在 device 进程启动时保证：
  - `survey_campaign`：`id`，`enabled`（0/1），`start_at` / `end_at`（unix 秒），`intro`（说明文案），`created_at`。
  - `survey_option`：`id`，`campaign_id`，`text`，`sort`，`enabled`。
  - `survey_response`：`id`，`campaign_id`，`wx_id`，`device_no`，`baby_name`，`option_snapshot`（JSON 数组，元素为提交时的选项文案），`comment`，`submitted_at`。唯一键 `(campaign_id, wx_id)`。
- **理由**：与预约表同一部署方式，不另起迁移工具。
- **备选**：只放配置文件 → 无法按用户留结果，拒绝。

### D2：同一时刻只服务「最新一条活动」给 App；Hub 可翻旧结果

- **选择**：App `current` 只看 `id` 最大的活动。Hub 保存当前活动的开关、时间、说明与选项行。Hub 提供「开启新一轮」：插入新 `survey_campaign`（默认关），旧活动提交保留，结果列表可按活动筛选，默认当前活动。
- **理由**：「该次问卷」对应活动 id。改文案仍是同一次；再收一轮才换 id。
- **备选**：改任一选项就换 id → 运维改错字会让已答用户再被弹，拒绝。

### D3：展示判定

- **选择**：`show=true` 当且仅当：存在活动、`enabled=1`、`start_at <= now <= end_at`、该 wxId 无提交行。`now` 为 unix 秒。Hub 保存的开始/结束按 **Asia/Shanghai** 墙钟换成 unix 秒。`enabled=0` 时忽略时间窗。起止缺失或 `end_at < start_at` 时 Hub 保存失败。无启用选项时仍可 `show=true`（允许只写意见）。
- **理由**：与已确认的开关优先、上海时区两点、空选可提交一致。
- **备选**：服务端再查喂养历史 → 跨 history 库，且客户端已有本地缓存门闩，拒绝。

### D4：提交

- **选择**：`POST` body 含 `campaignId`、`optionIds`、`comment`。`optionIds` 与 trim 后的 `comment` 都空则 4xx，不写行。选项必须属于该活动且 `enabled=1`，写入时把文案抄进 `option_snapshot`。`baby_name` 按会话 `deviceNo` 读本域 `user`，没有则空串。`wx_id`、`device_no` 只取网关注入的内部头，不信 body。已有唯一行则返回成功且不插第二行。活动不可展示（关、过期、id 不是当前）时拒绝新插入。
- **理由**：快照避免事后改文案污染旧结果；重复提交不把人重新算成未答。
- **备选**：只存 option id → Hub 改文案后历史对不上，拒绝。

### D5：App 路径、鉴权、统计

- **选择**：
  - `GET /device/app/api/survey/current` → `{ show, campaignId, intro, options: [{ id, text }] }`；`show=false` 时其余可为空。
  - `POST /device/app/api/survey/submit`。
  - `g.Meta` 放在 `api/v1` 新文件，不改旧结构。
  - `installDeviceProxyMiddleware` 增加 `/device/app/api/survey/*`。
  - 不加入 Bearer 白名单。
  - `maintenance_skip.go` 精确加入这两条 `METHOD + path`。注释写明负责人确认不统计。
- **理由**：与 appointment 反代方式一致；统计结论已确认。

### D6：Hub

- **选择**：新静态页 `/device/admin/survey`（文件 `survey-admin.html`），登记 `admin_static_pages.go` 与 `admin-modules.js`（`showInNav: true`）。加载 `admin-pages.css`。宽屏 `.hub-md`：左为开关、上海时区起止、说明、选项行编辑与「开启新一轮」；右为提交表（wxId、宝宝昵称、设备号、选项快照、意见、时间）。Admin API 落在 `/device/admin/api/survey/*`，已有反代前缀，不再单独 Bind。
- **理由**：符合 Hub 近全宽与左右分栏约定。

### D7：查询写法

- **选择**：可能无行的单行读取用 `.One()` + `IsEmpty()`。列表用 `Scan(&[]T)`。
- **理由**：遵守 GoFrame 空结果约定。

## Risks / Trade-offs

- [客户端时钟与服务器不一致] → 展示窗口只在服务端用 `now` 判断，客户端不自行比时间。
- [就地改选项文案后，未提交用户看到新文案、已提交用户仍是旧快照] → 接受；要统一口径就开新一轮。
- [无宝宝画像时昵称为空] → Hub 仍展示 wxId 与设备号。
- [并发双击提交] → 唯一键兜底，后者视为已提交成功。

## Migration Plan

- 部署 device-service 后首次启动建表。回滚：停用 Hub 开关即可让 App `show=false`；表可留。
- gateway-app 需同步带上反代 pattern 与 skip 列表，否则 App 到不了 device 或被误统计。

## Open Questions

- （无）
