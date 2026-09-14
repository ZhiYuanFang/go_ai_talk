## 1. 读路径与种子去掉 env/yaml 回退

- [x] 1.1 `EnsureSchema` 种子 `vip_monthly_19` 时 `apple_product_id` 固定空串；停止读取 `CASH_APPLE_PRODUCT_ID` 与 `cash.appleProductId`；保留「空值不覆盖已有 Apple ID、不覆盖已有 price_fen」
- [x] 1.2 `GetActiveProduct` 仅返回库内 `apple_product_id`，删除 env 回退
- [x] 1.3 Apple 验单期望 productId 仅取 `vip_product`；空则拒绝并提示在开通功能管理配置；更新 VIP 建单空 ID 时的 `payTip` 文案
- [x] 1.4 `config.cash-service.yaml` 去掉 `cash.appleProductId` 及「可被 CASH_APPLE_PRODUCT_ID 覆盖」注释；注释改为 Admin 配库

## 2. VIP 商品 Admin API

- [x] 2.1 `api/v1` 增加 `GET`/`POST /cash/admin/api/vip/product` 的 `g.Meta` 请求/响应类型（字段：title/priceFen/originalPriceFen/durationDays/appleProductId/status；productCode 只读 `vip_monthly_19`）
- [x] 2.2 cash-service 实现读写 `vip_product` 该行：校验 Admin 口令、price 非负、禁止插入第二行；controller 装配
- [x] 2.3 确认走既有 `/cash/admin/api/*` 反代；不改 `maintenance_skip.go`；无新增 App 路径故无需 Bearer 白名单

## 3. 开通功能管理 · VIP 套餐与 VIP列表

- [x] 3.1 `cash-feature-admin.html` 增加独立「VIP 套餐」表单（现价/原价/时长/Apple 商品 ID，单位分），加载 GET、保存 POST Admin VIP 商品接口
- [x] 3.2 同页增加按钮，可见文案精确为「VIP列表」，跳转 `/device/admin/cash-vip-admin.html`
- [x] 3.3 功能定义列表不得出现 VIP 假功能行
- [x] 3.4 `cash-vip-admin.html` 增加返回开通功能管理的链接；页内保持只读权益列表

## 4. Hub 导航

- [x] 4.1 `admin-modules.js` 中 `cash-vip-admin` 设 `showInNav: false`，保留 id 与 pagePath
- [x] 4.2 确认 `RegisterAdminStaticPages` 与静态页 Bearer 白名单仍含 `cash-vip-admin.html` / `cash-feature-admin.html`

## 5. 部署模板与 runbook

- [x] 5.1 `manifest/docker/.env.example` 删除 `CASH_APPLE_PRODUCT_ID`；保留 `CASH_APPLE_BUNDLE_ID`
- [x] 5.2 `docker-compose.microservices.yml` cash-service 去掉 `CASH_APPLE_PRODUCT_ID` 注入
- [x] 5.3 更新 `docs/runbooks/vip-commercial-config.md`、`cash-vip-sandbox.md`、`release-deploy-and-run.md`：Apple 商品 ID 改在开通功能管理配置；Hub 入口改为该页「VIP列表」；删除 env 字段说明

## 6. 自检

- [x] 6.1 仓库内检索 `CASH_APPLE_PRODUCT_ID`：业务读路径与 compose/example 应为 0（文档历史变更目录除外）
- [x] 6.2 人工：Admin 改 VIP 价与 Apple ID → App `GET /vip/product` 与建单金额/appleProductId 一致；Hub 无「VIP 权益」导航；「VIP列表」打开权益页
- [x] 6.3 确认未改 App v1 结构、未把 VIP 写入功能表、未新增测试文件、未加 Redis
