## Why

运维需要按活动收集「胖宝 App 哪些地方让用户满意」，并能开关、设有效期、查看每人的选择与意见。问卷配置与提交结果属于设备域用户画像，应由 device-service 提供接口与 Hub 页，而不是散落在客户端本地。

## What Changes

- device DB 新增满意度问卷活动、选项行、按用户提交结果；同一活动同一 wxId 只保留一条提交。
- device-service 新增两条需登录的 App API：读取当前是否展示、提交选择与意见。不修改既有 v1 接口结构。
- 展示条件：活动开关打开，且当前时刻落在上海时区的开始、结束时间点之内，且该 wxId 尚未提交本次活动。开关关闭时，即使仍在有效期内也不展示。
- 提交允许一个选项都不选、只写意见；选项与意见都空则拒绝，且不记为已提交。提交时快照选项文案、宝宝昵称、设备号。
- Hub 新增「问卷调查」页：左侧编辑开关、有效期与每一行选项；右侧查看提交（用户 id、宝宝昵称、设备号、所选结果、反馈）。宽屏左右分栏、近全宽。
- 新开一轮活动使用新活动 id，旧提交仍可按活动查看；就地改文案不重置已提交状态。
- **usage**：负责人已确认 `GET /device/app/api/survey/current` 与 `POST /device/app/api/survey/submit` **不计入**功能使用统计，写入 `maintenance_skip.go` 精确排除。Hub `/device/admin/api/*` 沿用既有结构性跳过。
- Flutter 弹窗不在本仓实现，契约供兄弟仓 `home-satisfaction-survey` 消费。

## Capabilities

### New Capabilities

- `device-satisfaction-survey`：满意度问卷的活动配置、App 展示/提交契约、一次提交关闭、Hub 编辑与结果查看。

### Modified Capabilities

- （无）

## Impact

- **device-service / device DB**：新表与 DAO；App 与 Hub admin handler；中文业务注释。禁止跨库直查他域表。宝宝昵称取本域 `user.baby_name`（按提交时会话 deviceNo）。
- **gateway-app**：`installDeviceProxyMiddleware` 增加 `/device/app/api/survey/*`（`/device/admin/api/*` 已覆盖 Hub API）。Bearer **不**豁免。`maintenance_skip.go` 精确排除上述两条 App 路径。`api/v1` 的 `g.Meta` 供 apiregistry。
- **Hub 静态页**：`resource/public` 新页、`admin_static_pages.go`、`admin-modules.js` 导航；加载 `admin-pages.css`，`.hub-md` 左右分栏。
- **Redis**：不为问卷配置或结果新增读缓存。
- **Flutter**：兄弟仓另案实现；本仓只保证 HTTP 契约。
