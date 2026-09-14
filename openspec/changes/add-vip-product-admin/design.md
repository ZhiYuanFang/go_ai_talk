## Context

VIP 与商业功能开通同在 `cash-service` / `ai_voice_cash`，但商品配置入口分裂：

- 功能 SKU：`feature_product`，开通功能管理可改 `price_fen` / `apple_product_id`。
- VIP：`vip_product` 一期固定 `vip_monthly_19`；现价靠 SQL；Apple 商品 ID 还可被 `CASH_APPLE_PRODUCT_ID` / yaml `cash.appleProductId` 在读路径与 EnsureSchema 种子上回退。Hub「VIP 权益」页只读 `vip_entitlement`，不能改商品。

App 侧 VIP 仍是账号 overlay（`GET /cash/app/api/vip/*`，履约写 `vip_entitlement`），**禁止**写入功能权益表。本变更只把 VIP **商品配置**收进开通功能管理，并收掉 env 回退。

约束：不改已发布 App v1 结构；Admin 走既有 `/cash/admin/api/*` 反代与口令注入；不加 Redis；不新建微服务；不新增测试文件。

## Goals / Non-Goals

**Goals:**

- 运维在开通功能管理配置 VIP 现价/原价/时长/Apple 商品 ID。
- 建单、匿名读价、Apple 验单只认 `vip_product` 行，无 env/yaml 回退。
- 部署模板删除 `CASH_APPLE_PRODUCT_ID`；保留 `CASH_APPLE_BUNDLE_ID`。
- Hub 主导航不再出现「VIP 权益」；开通功能管理用「VIP列表」打开既有只读权益页。

**Non-Goals:**

- 将 VIP 做成 `feature_def` 第四项或出现在 App catalog。
- 多档 VIP SKU、退款、手工开通、ASC 自动同步。
- 支付宝密钥进 Admin 或改 `CASH_ALIPAY_*`。
- 改 Flutter 契约或 usage 统计 denylist。

## Decisions

### D1：VIP 商品仍在 `vip_product`，不并入 `feature_product`

- **选择**：Admin 读写 `vip_product` 唯一行 `vip_monthly_19`。开通功能管理增加独立「VIP 套餐」区块，列表不混入功能定义。
- **理由**：overlay 履约、匿名 `/vip/product`、care-alert/成长轨迹 `isVip` 短路均依赖 VIP 表。并入功能表会污染 catalog 与邀请码。
- **备选**：`feature_id=vip` 隐藏目录项 — 履约与 App 分流成本高，否决。

### D2：Admin 契约 `GET` + `POST /cash/admin/api/vip/product`

- **选择**：与功能 SKU 相同鉴权（网关注入 `X-Admin-Password`）。GET 返回一期商品字段（`productCode`、`title`、`priceFen`、`originalPriceFen`、`durationDays`、`appleProductId`、`status`）。POST 更新上述可写字段；`productCode` 只读，MUST 为 `vip_monthly_19`（省略则按该码更新）。不允许新建第二行 VIP SKU。
- **理由**：App `GET /cash/app/api/vip/product` 已占用匿名读价路径，Admin 用 `/cash/admin/api/` 前缀避免混淆；不新增顶层反代。
- **备选**：复用 `POST /cash/admin/api/feature/products` 特判 featureId=vip — 语义混乱，否决。

### D3：读路径取消 Apple 商品 ID 回退

- **选择**：`GetActiveProduct` / Apple 验单只使用 `vip_product.apple_product_id`。空则 iOS 建单 `appleProductId` 为空并带明确 `payTip`；验单在期望 ID 为空时拒绝（提示走开通功能管理配置）。EnsureSchema 种子 `apple_product_id=''`，**停止**读取 `CASH_APPLE_PRODUCT_ID` 与 `cash.appleProductId`。`ON DUPLICATE KEY UPDATE` 保持：空值不覆盖已有 Apple ID；**不**用种子覆盖已有 `price_fen`。
- **理由**：单一真相源，避免「env 有值、库为空、Admin 改了但读路径仍吃 env」。
- **备选**：保留 env 覆盖 Admin — 与「去掉该字段」冲突。

### D4：Hub 入口

- **选择**：`admin-modules.js` 中 `cash-vip-admin` 设 `showInNav: false`，保留 `id`/`pagePath` 以便静态注册。`cash-feature-admin.html` 主按钮区增加 **「VIP列表」**（`secondary`），`href`/`location` 到 `/device/admin/cash-vip-admin.html`。权益页增加返回开通功能管理的链接。`RegisterAdminStaticPages` 与 Bearer 静态白名单不变。
- **理由**：用户明确要挪入口且文案为「VIP列表」；只读列表与商品编辑职责分离，不必把表格塞进同一超长页。
- **备选**：iframe 嵌入权益表 — 无必要。删除静态页 — 丢失查谁开过 VIP。

### D5：VIP 套餐表单字段

与功能套餐对齐、单位为分：现价、原价（0=不划线）、授予天数、Apple 商品 ID；标题默认可编；上架状态保留但一期默认保持上架。保存走 D2 POST。价格预览可复用功能页分→元。

### D6：Redis / usage / App 版本

- 不加 Redis：运维写极少，App 读价已直查 MySQL。
- 无新增 App HTTP：不改 `maintenance_skip.go`。
- App v1 结构不变；仅 Admin 新增路径（`g.Meta` 供 apiregistry）。

## Risks / Trade-offs

- [生产从未在库中写入 Apple ID，只靠 env 读时回退] → 当前 prod/test 均未配 `CASH_APPLE_PRODUCT_ID`；部署后须在 Admin 填写。若某环境曾只配 env、EnsureSchema 已写入库，则删除 env 后库值仍有效。
- [运维误把 VIP 改价当成 ASC 同步] → 文案标明 ASC 标价须人肉对齐；本服务不调 Apple 后台。
- [Hub 收藏旧「VIP 权益」书签] → 静态 URL 仍可用；主导航隐藏不影响直达。

## Migration Plan

1. 发版 cash-service + gateway-app 静态资源。
2. 从 compose / `.env.example` 去掉 `CASH_APPLE_PRODUCT_ID`（本地 `.env.prod` 若曾手写该键可删，无功能影响）。
3. 运维打开开通功能管理 → 填 VIP 价与 Apple ID → 保存；用「VIP列表」确认权益页仍可用。
4. 回滚：还原镜像即可恢复 env 回退（若旧镜像仍读 env）；Admin 写入的库字段可保留。

## Open Questions

无。方案 A 与「VIP列表」文案已由产品确认。
