## Context

现网：值得留意喂养资格在 cash（`care_alert_entry`，默认连续 2 有效日）；Flutter 先查 eligibility 再调 voice `CareAlertDaily`。voice 侧仅校验 `wxId`/`deviceNo`，**无功能开通门禁**；VIP 只影响选模/额度，不决定「能否看」。

商业开通已有 `feature_def` / `feature_entitlement`（按 `device_no`）/ `feature_allowed_count`（预测条数）/ 支付·邀请·广告分支，但：

- 邀请对非预测硬传 `durationDays=0`（永久），忽略 `feature_def.duration_days`；
- Admin 功能定义页未暴露邀请授予天数；
- 履约逻辑散落 if-预测 / if-其它，缺主体与通道枚举的统一入口；
- 无「设备对本功能仅兑一次邀请」闸门。

约束：跨服务禁止直查他域库；Redis 经 `cachekit`；gateway-app 对外接口须反代/usage/白名单自检；不改已有 App API 结构版本字段（新增路径用既有 v1 风格）；本阶段不新增测试文件。

## Goals / Non-Goals

**Goals:**

- 共用开通原子工具，按主体（user|device）与通道（payment|invite_code|ad）入参驱动授予。
- 值得留意独立 `featureId`：付费永久；邀请/广告限时可配；device 兑码一次；VIP 覆盖使用权。
- cash internal「可看值得留意」合成；voice dual-gate fail-closed。
- Admin 可配邀请/广告授予天数；Flutter 合格未开通可进开通中心。

**Non-Goals:**

- 改变预测 `allowed_count` 累加或 VIP 不写条数语义。
- 用户维（`user` subject）权益落表一期落地（枚举预留，非本功能路径）。
- 广告验真、Flutter 强制广告 CTA。
- 改喂养资格算法或允许 VIP 短路 `qualified`。
- 新建微服务 / 新测文件。

## Decisions

### D1 — 原子工具 `ActivateFeature`

单一入口（名称以实现为准）接收：

| 字段 | 语义 |
|------|------|
| FeatureID | 稳定功能编号 |
| SubjectType | `device` \| `user`（一期值得留意/预测均 `device`；`user` 调用拒绝或 no-op 明确错误） |
| SubjectKey | `device_no` 或未来 `wx_id` |
| Channel | `payment` \| `invite_code` \| `ad` |
| ChannelRef | 订单号 / 邀请码 / 广告幂等键 |
| ActorWxID | 操作者（兑码用户）；原力与流水用；可与 Subject 分离 |
| GrantKind / Qty | 可由 feature 策略覆盖；预测强制 count delta |

效果解析：

- **payment**：读 SKU/`feature_product.duration_days`（值得留意种子 0=永久）；grant_kind 来自 SKU。
- **invite_code / ad**：**同源**读 `feature_def.duration_days`（0=永久，>0=自授予起 N×86400）；预测类忽略天数、走 `allowed_count_delta +1`。
- 支付回调、`RedeemInviteCode`、`CompleteFeatureAd` MUST 调用该入口，禁止旁路再写表。

**备选**：继续在三处复制 if — 否决（与「后续更快接入」冲突）。

### D2 — 值得留意 `featureId` 与种子

- 稳定 ID：`care_alert_smart_remind`（客户端常量对齐）。
- `unlock_methods`：`payment,invite_code,ad`。
- `duration_days` 种子默认 **7**（邀请/广告）；Admin 可改。
- `grant_kind` 权益型；SKU 付费 `duration_days=0` 永久、`grant_quantity=1`。
- EnsureSchema 幂等 INSERT，不覆盖运维已改 `duration_days`/文案。

### D3 — 邀请：设备主体 + 用户流水

```
兑码成功
  ├─ ActivateFeature(subject=device, channel=invite)
  ├─ 人×码×功能去重（既有）
  ├─ 本功能额外：device×feature 邀请成功仅一次
  └─ 原力 NotifyUcg → owner 用户（既有）；redemption 记 redeemer_wx_id
```

新增去重存储（表名以实现为准，如 `feature_invite_device_grant`）：`PRIMARY KEY (device_no, feature_id)`，仅对配置了「设备邀请一次」的功能写入（至少 `care_alert_smart_remind`）。预测 MUST NOT 使用该闸门。

同宝宝多账号：A 兑码开设备权益，B 共享——符合全家共享。

### D4 — cash internal access + voice 门禁

```
GET /cash/internal/api/care-alert/access?deviceNo=&wxId=
→ {
    allowed: bool
    feedingQualified: bool
    featureActive: bool   // 未过期 device entitlement OR isVip(wxId)
    // 可选：reason、feeding 进度字段、entitlementExpiresAt
  }
```

合成顺序：先喂养资格（复用 `GetCareAlertFeedingEligibility`），再 `featureActive`；`allowed = feedingQualified && featureActive`。

- VIP 查询失败：当作非 VIP（仅认 entitlement）；整体 cash HTTP 失败：voice **fail-closed** 拒绝业务（打 Warning）。
- voice：`CareAlertDaily` 每次请求（含缓存命中）调用 access；未 `allowed` 不得返回内容/不得生成。Delete/Feedback：至少 Daily 强制；删除/反馈若无日缓存可同样校验或仅登录校验——**MUST 对 Daily 强制**；Delete/Feedback SHOULD 校验 `featureActive`（防过期后改数据），喂养资格对写操作可放宽（设计默认：三条均调 access，与 Daily 一致，避免旁路）。

客户端 Remote 封装对齐 `RemoteIsVipByWxID`（`CASH_SERVICE_URL` + internal secret）。

**备选**：voice 分别调 eligibility + VIP + 自实现 entitlement — 否决（多次往返、语义易分叉）。

### D5 — VIP 覆盖

与 `vip-overlay-entitlement` 一致：VIP 履约不写 `feature_entitlement`；catalog 返回真实设备开通态；客户端 `isFeatureEffectivelyUnlocked`：非预测项 `isVip || unlocked`；access 内 `featureActive` 含 VIP。

### D6 — Admin UI

开通功能管理「功能定义」增加「邀请/广告授予天数」字段，绑定 `feature_def.duration_days`；副文案说明：0=永久；付费看套餐；预测条数类忽略。SKU 区不变。

### D7 — Flutter 三态

| 态 | UI |
|----|-----|
| 喂养未合格 / 资格失败 | 进度或失败文案（现状） |
| 合格且未开通且非 VIP | 引导文案；点击 → `FeatureUnlockHub`（可带 featureId 锚点） |
| 合格且（开通或 VIP） | 拉 daily / 跑马灯 |

客户端可先本地拼态，但服务端 access/Daily 为权威。错误码：若 access/App 暴露 reason，客户端宜区分两门槛。

### D8 — usage / 网关

- internal access：进程间，不计 App usage。
- 若新增 App 侧合成读接口：须先问负责人是否计入 usage；未答复前不改 `maintenance_skip` 假定策略。既有 eligibility / catalog 查询 skip 不变；开通 POST 仍统计。
- 不加入 Bearer 匿名白名单。

### D9 — Redis

- 不默认新读缓存；access 可复用喂养资格日缓存键族；设备权益读以 MySQL 为准，目录缓存失效沿用既有 `invalidateDeviceFeatureCaches`。
- 若 access 需独立短缓存：须负责人确认后再加 `cachekit` 键 builder。

## Risks / Trade-offs

- [voice 强依赖 cash] → 超时/失败 fail-closed；监控 Warning；超时预算 ≤ VIP 查询同级（约 5s 客户端已有）。
- [日缓存命中仍打 access] → 略增 QPS；换取邀请过期立刻失效；可接受。
- [device 兑码一次过严] → 仅值得留意；换设备可再兑；同设备换账号不可再兑（产品意图）。
- [重构邀请/广告汇入 Activate] → 回归预测 +1 与现有兑码；任务中显式自检预测路径不变。
- [Flutter 跨仓] → 本仓规格约束孪生；合并前对齐 featureId 常量。

## Migration Plan

1. 部署 cash：EnsureSchema 种子功能 + 设备邀请去重表；Activate 汇入；Admin 字段；internal access。
2. 部署 voice：Daily（及 Delete/Feedback）调 access。
3. 发 Flutter：三态 + 开通中心展示新目录项。
4. 回滚：回退 voice 即去掉服务端开通闸（仅喂养仍可由客户端拦）；cash 种子行可保留无害。

## Open Questions

- access 是否向 App 暴露同构只读接口（便于 Flutter 少拼态）——一期可用 eligibility + catalog + isVip 本地合成；若加 App API 须问 usage。
- `user` subject 表结构二期再定；一期枚举拒绝即可。
- Delete/Feedback 是否与 Daily 完全同 access（本设计默认是）。
