# Design

## Context

See `proposal.md` — Why / What Changes。现网约束：

- C 端身份以 `wx` + 单列 `wx.device_no` + App JWT `device_no` claim 为中心；Flutter 旧版仅「无→一」绑定。
- 商业化在 `cash-service`：VIP/功能为一次性支付 + 时长续期，**无**商户连续订阅轨。
- 历史实时：`history-service` → Redis → gateway App SSE/WS（一连接一 `device_no`）。
- 运维 Hub 为 `resource/public` 静态 HTML；不适合作商户玻璃拟态 SPA。
- 兄弟仓 `flutter_ai_talk` 已有玻璃卡/预测/趋势可对标；本 change **同发**新版多宝宝 UI。

跨库禁直查；商户权益在 cash，成员/商户归属在 device，写闸在 history 经内部探活。

## Goals / Non-Goals

**Goals:**

- 独立 `merchant` 身份与 JWT，与 wx 会话严格隔离。
- 成员表多宝宝真源 + `wx.device_no` 焦点镜像，旧 App 可观察行为不变。
- 商户配额（2 / 不限量）、分发/解绑、欠费 create 闸、Hub 绝对授期与支付宝连续订阅共用 `expire_at`。
- 商户 Vue SPA + 单连接多 device SSE；Flutter 新版多宝宝与 parent 创建/绑定虚拟宝宝。

**Non-Goals:**

- 不把商户登录并入 Hub Admin 口令体系。
- 不在一期做 parent 自助解绑或「自然人跨多 wx 行」防刷。
- 不要求旧 App 支持多宝宝切换或再绑。
- 不将商户连续订阅履约写入 C 端 `vip_entitlement`。
- 不在本设计锁定支付宝签约产品的每一字段名（实现时对照开放平台文档，行为以 specs 为准）。

## Decisions

### D1 — 商户独立表 + 独立 JWT（非 wx.role=merchant）

- **选择**：`merchant` 表（店名唯一、密码哈希）+ gateway 签发商户 access；子账户才是 `wx`。
- **理由**：登录面、令牌 claim、路由隔离最清晰；避免商户 JWT 误注入 `X-Internal-Wx-Id`。
- **备选**：wx 加 merchant role — 否决（与「只能登后台」冲突面更大）。

### D2 — 成员表 + 焦点镜像列

- **选择**：`wx_baby_member(wx_id, device_no, ...)` 为真源；`wx.device_no` = 焦点；JWT 仍带当前焦点。
- **迁移**：上线任务将现有非空 `wx.device_no` 回填成员表一行。
- **备选**：废弃镜像列只靠 JWT — 否决（破坏 `detail`/旧客户端）。

### D3 — role 账号单值 parent|staff 互斥

- **选择**：列或等价字段；商户创建子账户 = username 注册 + `role=staff`；禁止 parent↔staff 变更。
- **staff 宝宝来源**：一期以商户分发为主；parent 自建/扫码仅 `role=parent`。

### D4 — 宝宝归属商户

- **选择**：`device`/`user` 侧记录 `owner_merchant_id`（或 `merchant_baby` 归属表）。写闸与配额都查归属→cash 权益。
- **画像**：后台 PATCH 与 App 画像 API 写同一画像行，后写覆盖。

### D5 — 欠费写闸位置

- **选择**：history 权威 create 路径调用 cash/device 内部「该 device 是否因商户欠费禁止 create」；update/end/读放行。
- **理由**：覆盖 Python batch 与 App，避免只拦 HTTP 网关。

### D6 — 权益单字段 expire_at

- **选择**：`merchant_entitlement.expire_at`；Hub 写**绝对时间**；支付宝连续订阅成功按周期延长同一字段（`max(now, expire_at)+period` 或回调约定，须幂等）。
- **默认商品**：1000 元/月，Hub 可改 `price_fen`。
- **备选**：Hub 与支付宝两套到期 — 否决（与产品决定冲突）。

### D7 — 商户 SSE

- **选择**：复用 `app:history:notify`（或现网约定 channel）；gateway 维护 `merchant_id → set(device_no)`，过滤后下行加 `deviceNo`。
- **集合更新**：创建/分发/解绑归属后 bump 版本或重载 set。

### D8 — 前端同发

- **Vue**：独立工程目录（建议 `web/merchant-console` 或 `resource/merchant-spa` 构建产物），玻璃拟态对标 Flutter tokens。
- **Flutter**：同发多宝宝切换；创建/绑定虚拟宝宝；无解绑 UI。
- **旧 App**：仅焦点宝宝；staff 多成员时分发指定默认焦点。

### D9 — 单一大 change

- tasks 分泳道：identity / commerce / history-gate / gateway-sse / hub / vue / flutter / deploy。
- 归档后将 `merchant-console` 纳入主规格；`project.md` 域表从十七域增补该域。

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| 支付宝连续订阅对接周期长 | Hub 绝对授期可先支撑内测店；订阅轨并行，共用 `expire_at` |
| staff 多宝宝 × 旧 App 只见 1 | 产品已接受；分发时设默认焦点；新 App 同发切换器 |
| 成员表与镜像不一致 | 写成员/切焦点同一服务路径更新；迁移校验脚本 |
| history 写闸增加延迟/依赖 | 短超时 + 失败策略显式（建议：探活失败时对「已知商户宝宝」拒绝 create，对非商户放行——实现须在任务中固定） |
| Vue 与 Hub 两套前端 | 受众不同；共享设计 token 文档而非代码强耦 |
| Flutter 同发协调 | tasks 标明兄弟仓清单；gateway 契约先冻结再改 UI |

## Migration Plan

1. DDL：`merchant`、成员表、`wx.role`、归属、`merchant_entitlement`、商品种子（1000/月）。
2. 回填：`wx.device_no` → 成员一行；`role=parent` 默认。
3. 部署：cash → device → history → gateway；再发 Hub 页与商户 SPA；再发 Flutter。
4. 回滚：功能开关关闭商户路由与写闸；成员表保留不影响旧读路径（焦点列仍在）。

## Open Questions

- 支付宝具体产品（周期扣款 vs 商家扣款）字段级对接表 — 实现前 spike，不改变「连续订阅 + 共用 expire_at」行为。
- 商户 SPA 仓库内路径最终命名 — 任务中选定不影响规格。
- parent 创建第二个宝宝后默认焦点保持或切换 — 实现时定一并在 App 文案一致；规格允许文档化任一策略。
