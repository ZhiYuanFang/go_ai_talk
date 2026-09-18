## Why

开通功能管理里保存售卖套餐时，商品编码为空就会新建一条随机编码。App 与 Apple 只认种子编码，新编码会报商品不存在。改价、改天数应当改原套餐。

## What Changes

- Admin 保存功能 SKU 时，商品编码为空，或库中不存在该编码，MUST 拒绝，MUST NOT 插入新行。已有编码只更新价格、天数、Apple 商品 ID、上下架等字段，不改编码。
- 打开某个功能时，表单填入该功能已有套餐（优先上架且非 `fp_` 自动生成码）。没有套餐时不生成编码。
- 去掉「新建套餐（清空编码）」。
- 不批量下架、不删除已经误建的 `fp_` 行，也不改历史订单。

## Capabilities

### New Capabilities

- `feature-sku-update`: 功能售卖套餐只能更新已有商品编码，管理页打开时填入原套餐。

### Modified Capabilities

- （无）不改 App 下单字段、不改 VIP 套餐保存。

## Impact

- **进程**：`cash-service`。不新增表，不新增配置组。仍用 `default` / `CASH_DB_LINK`。
- **代码**：`internal/services/cash/feature_admin.go` 的 SKU 保存；`resource/public/cash-feature-admin.html`。
- **不做**：不改 App API 结构，不改 `maintenance_skip`，不加 Redis，不加后台任务，不改网关 Bind。
