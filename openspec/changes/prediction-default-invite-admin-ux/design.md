## Context

`commercial-feature-entitlement` 已落地：`feature_def` / `feature_product` / 邀请码 / `feature_allowed_count` / 合成 catalog。当前缺口：

- 无设备记录时 `allowedCount=0`，无全站默认免费条数。
- 邀请码 `grant_duration_days` 对预测路径无效（`allowed_count_delta` 只累加、不写到期）。
- Admin 手填 `featureId`、商品编码、逗号串开通方式，易与客户端约定冲突。

本设计只改 `cash-service` + Admin 静态页 + 必要的 `api/v1` 字段；不新建服务、不加 Redis 新缓存族（沿用既有 allowedCount/定义缓存，履约后失效）。

## Goals / Non-Goals

**Goals:**

- 有效条数 = **默认 + 永久累加（A）**；默认可 Admin 配置。
- 邀请码兑换预测 = **T1 临时全开**；期限内 catalog `allowedCount=-1`，并带到期时间；到期回落 A。
- Admin：功能编号只读；标题/简介可编；SKU 编码自动生成；所属功能下拉；开通方式多选；按钮纵向对齐。

**Non-Goals:**

- 改 Flutter（仅文档约定哨兵）；广告服务端验真；邀请分润；UCG/VIP 语义；新建测试文件；新微服务/ACR。

## Decisions

### D1 — 有效条数公式（A）

- `permanentDelta`：`feature_allowed_count.allowed_count`（仅付费/广告等永久来源写入；邀请码**不再**对该字段做预测累加）。
- `defaultFree`：存 `feature_def` 扩展字段，如 `default_allowed_count`（仅 `prediction_unlock` 有意义；其它功能可 0）。
- 读路径：`effective = defaultFree + permanentDelta`（非负；无临时全开时）。

**备选**：配置 yaml 默认条数 → 否决（运维要在 Admin 改，不走发版）。

### D2 — T1 临时全开存储

在 `feature_allowed_count` 增加 `full_access_expires_at BIGINT NOT NULL DEFAULT 0`：

- 邀请码兑换 `prediction_unlock` 且 `grant_duration_days>0`：`full_access_expires_at = max(现有未过期值, now) + days*86400`（与权益续期类似）；`grant_duration_days=0` 表示永久全开则写 `0` 哨兵约定：**0 + 需区分「无临时」与「永久全开」**。

永久全开约定：

- `full_access_expires_at = 0` 且 **无**「已启用临时全开」标记 → 无临时层。
- 为消歧：增加 `full_access TINYINT`：`0=无`，`1=有`；有则 `expires_at=0` 表示永久全开，`>now` 表示期限内。

履约：邀请码预测只更新 `full_access` / `full_access_expires_at`，**不**增加 `allowed_count`。

Catalog：

- 若 `full_access=1` 且（`expires_at=0` 或 `expires_at>now`）：`allowedCount=-1`，`expiresAt` 在有限期时填 unix；永久全开可不填或 0。
- 否则：`allowedCount = defaultFree + allowed_count`。

**备选**：用 `feature_entitlement` 表示临时全开 → 可行但 catalog 已把预测当 count 项，双读更绕；优先扩 `feature_allowed_count`。

### D3 — 哨兵 `-1`

客户端约定：`allowedCount == -1` → 全部预测事项可看。正整数 → 前 N 条。不引入额外布尔字段（产品已选哨兵）。

### D4 — Admin 功能编号只读

- 禁止 Admin「新建任意 featureId」；列表/编辑：`featureId` 只读。
- 新功能仅靠 EnsureSchema 种子或受控发版脚本。
- Upsert API：若 `feature_id` 不存在 → 拒绝（或仅允许更新已存在行）；已存在则禁止改主键。

### D5 — 商品编码自动生成

- 新建 SKU：`product_code` 空则服务端生成（如 `fp_{featureId缩写}_{unix}_{短随机}`），保证唯一。
- 更新：必须带已有 `productCode`，禁止改码。
- Admin UI：新建不填编码；保存后列表展示可复制。

### D6 — 开通方式多选

- UI：三枚举 checkbox/多选；保存 join 为现有逗号串。
- App/API：`unlockMethods` 字符串契约不变。

### D7 — 邀请码与「开通多少」

- 预测邀请：不再用 `grant_quantity` 表达「开几条」；「多少」= 全开（`-1`），「多久」= `grant_duration_days`。
- 码仍可挂 `prediction_unlock`；兑换协议可仍带 `featureId=prediction_unlock`（兼容现网一次一功能）。

### D8 — Redis

- 沿用 `CashFeatureAllowedCountKey`；写入临时全开/默认变更后失效定义缓存与设备 allowedCount 缓存。不新增键族（若缓存 value 结构变，失效即可或 bump 逻辑版本）。

## Risks / Trade-offs

- [客户端未识别 `-1`] → 文档/联调清单；上线前对齐 Flutter。
- [历史设备曾被邀请码 +N] → 永久 delta 保留；新兑换改 T1，不回滚历史 delta。
- [永久全开 `expires_at=0` 与「无记录」混淆] → 用 `full_access` 标志消歧。
- [Admin 不能自助加新功能 ID] → 刻意约束；新功能走发版种子。

## Migration Plan

1. EnsureSchema：`feature_def.default_allowed_count`；`feature_allowed_count.full_access` + `full_access_expires_at`。
2. 种子：为 `prediction_unlock` 设合理 `default_allowed_count`（产品可给初值，缺省 0）。
3. 部署 cash + 刷新 Admin 静态资源；gateway-app 若托管静态页则一并发。
4. 回滚：代码回退；新列可留（兼容旧读忽略）。

## Open Questions

- `prediction_unlock` 的初始 `default_allowed_count` 具体数字（实现前向产品确认，缺省 0 可上线）。
- 邀请码 `grant_duration_days=0` 对预测是否允许「永久全开」：本设计允许；若产品禁止永久，Admin 校验 `>0` 即可（实现时可加可选校验）。
