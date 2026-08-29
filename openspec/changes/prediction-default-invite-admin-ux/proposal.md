## Why

商业功能开通已上线合成 catalog 与邀请码履约，但预测 `allowedCount` 缺「全站默认条数」、邀请码对预测无法真正落地「有期限的临时全开」，且 Admin 把 `featureId`/商品编码/开通方式当自由文本，易与客户端约定脱节、运维易错。本变更在既有 cash 功能域上补齐有效条数语义与 Admin 契约化 UX，客户端继续直出服务端标题/简介。

## What Changes

- **预测有效条数（语义 A）**：`effectiveAllowedCount = defaultFree + permanentDelta`；`defaultFree` 可配置（功能定义或等价配置）；永久累加仍来自付费/广告等非临时来源。
- **邀请码预测（T1）**：兑换预测功能在码配置的授予有效天数内为**临时全开**；catalog 预测项 `allowedCount` 使用哨兵 **`-1`** 表示全开，并暴露到期时间（沿用/补齐 `expiresAt`）；到期后回落为语义 A（永久累加保留）。
- **BREAKING（客户端约定）**：`-1` 表示预测临时全开；正整数仍为可看条数上限。须与 Flutter 同步约定。
- **功能编号**：与客户端约定、种子/发版写入；Admin **禁止**新建随意编号、**禁止**修改已有 `featureId`（只读展示）；可编辑标题、简介、开通方式、上架、默认条数等展示/策略字段。
- **标题/简介**：继续经 catalog 下发，客户端直接展示；Admin 可完整编辑简介。
- **售卖套餐**：商品编码**服务端自动生成**（新建时），Admin 只读展示；所属功能改为**下拉选择**已有功能。
- **开通方式**：Admin 改为固定枚举**多选**（payment / invite_code / ad），禁止逗号手拼；持久化可仍为逗号串以兼容现网校验与 App 字段。
- **Admin 布局**：操作按钮与表单字段纵向对齐（开通功能页、邀请码页一致原则）。
- **不**新建微服务；**不**改 App usage 策略假设（查询仍 skip、开通意图仍计）；**不**要求改 VIP 写功能表语义。

## Capabilities

### New Capabilities

- `prediction-effective-count`：默认条数 + 永久累加（A）；邀请码 T1 临时全开；catalog 哨兵 `allowedCount=-1` 与到期回落。
- `feature-admin-contract-ux`：功能编号只读契约、文案可编、SKU 编码自动生成、所属功能下拉、开通方式多选、表单按钮纵向对齐。

### Modified Capabilities

- （无）`prediction-allowed-count` / `feature-admin-*` 等尚未归档入 `openspec/specs/`；行为增量以本变更 capabilities 为准，并与 `commercial-feature-entitlement` 对齐。

## Impact

- **进程**：`cash-service`（条数合成、邀请履约、Admin API/静态页）；`api/v1` catalog/Admin 请求响应字段（如默认条数、预测 `expiresAt`）。
- **库**：`ai_voice_cash`（如 `feature_def` 默认条数字段、`feature_allowed_count` 或权益表承载临时全开到期）；EnsureSchema 迁移兼容。
- **客户端**：须识别 `allowedCount=-1`；标题/简介仍用 catalog。
- **非目标**：新微服务；广告验真；邀请码分润；改 UCG 资格；改 VIP 履约写功能表；新增测试文件。
