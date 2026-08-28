## 1. Schema 与 Redis 键

- [x] 1.1 在 cash `EnsureSchema` 新增：`feature_def` / `feature_product` / `feature_invite_code` / `feature_invite_code_feature` / `feature_invite_redeemer_bind` / `feature_invite_feature_grant` / `feature_invite_redemption` / `feature_entitlement` / `feature_allowed_count` / `feature_order`（字段对齐 design D2）
- [x] 1.2 在 `internal/platform/cachekit/keys_*.go` 登记资格日缓存、功能定义缓存、可选 per-device 合成列表缓存、`allowedCount`、广告幂等键（中文注释：TTL/失效/共享语义）
- [x] 1.3 确认仅 `CASH_DB_LINK` / `ai_voice_cash`；不新建微服务与 ACR 矩阵项
- [x] 1.4 **已与负责人确认 Redis 策略**：加资格/定义/`allowedCount`/广告幂等（见 design）；结论已写入 design

## 2. History 按日统计契约

- [x] 2.1 history-service 新增内部接口：按 `deviceNo` 返回近 N 个上海日每日 history 条数；内部密钥校验；最少 DB 扫描
- [x] 2.2 cash 侧 HTTP 客户端经 `HISTORY_SERVICE_URL` + 内部密钥；禁止跨库；失败 fail-closed 可观测日志
- [x] 2.3 compose / `.env.example` / runbook 为 cash 补充 `HISTORY_SERVICE_URL`（及密钥惯例）；`config.cash-service.yaml` 不回流 history DB

## 3. UCG 入场资格

- [x] 3.1 连续有效日算法（≥10 条/上海日、连续 7 日、字段齐全）；**不读 VIP/功能表**
- [x] 3.2 Redis 按日缓存（miss 同步计算；不落 MySQL；经 cachekit）
- [x] 3.3 App `GET /cash/app/api/ucg/eligibility`：只信 `X-Internal-Device-No`；`api/v1` g.Meta

## 4. 合成功能目录与预测数量

- [x] 4.1 App `GET /cash/app/api/feature/catalog`：登录+绑机；项含 `unlocked` / 开通字段；预测项含 `allowedCount`；不含 UCG；只信内部设备头
- [x] 4.2 功能定义 Redis 热读 + Admin 写失效；设备权益请求路径 JOIN（可选 per-device 缓存）
- [x] 4.3 `api/v1` g.Meta；**不**实现 App 独立 entitlements / allowed-count 必调路径

## 5. 开通履约（支付 / 邀请码 / 广告）

- [x] 5.1 功能建单 `POST /cash/app/api/feature/orders`（`feature_order` + 渠道调起；绑定 device_no/wx_id）；计入 usage；g.Meta
- [x] 5.2 **共用**支付宝 notify / Apple verify：按订单表分流；VIP 只续期 VIP；功能写权益或 +allowedCount；幂等；VIP 路径不写功能表
- [x] 5.3 邀请码兑换 `POST /cash/app/api/feature/invite-codes/redeem`（`code`+`featureId`；不可自用；一家锁定仅成功后绑定；人×功能一次；同码可逐功能兑；流水）；计入 usage
- [x] 5.4 广告完成 `POST /cash/app/api/feature/ad/complete`（MVP 信客户端 + 幂等/日限额）；计入 usage
- [x] 5.5 确认既有 VIP status / internal VIP 读行为不变

## 6. Admin API 与 Hub UI

- [x] 6.1 付费功能 Admin CRUD：`/cash/admin/api/feature/**`（defs、products、解锁方式）
- [x] 6.2 邀请码 Admin API：CRUD、owner 一人一码、有效期、可开功能、按码兑换明细（wx_id/device/feature/time）
- [x] 6.3 静态页 `cash-feature-admin.html` + `cash-invite-code-admin.html`（AdminCommon + 可写 CRUD）
- [x] 6.4 `admin-modules.js` 登记两个入口；`RegisterAdminStaticPages` + 静态 path exempt

## 7. Gateway / usage / 文档自检

- [x] 7.1 确认 `/cash/app/api/*` 与 `/cash/admin/api/*` 反代覆盖；功能/资格路径 **不**加入 Bearer 匿名白名单；notify 保持既有 exempt
- [x] 7.2 **已与负责人确认 usage**：查询（eligibility、catalog）→ `maintenance_skip`；开通意图 POST（orders、invite redeem、ad complete）→ 统计；结论：查询 skip、开通 POST 统计（见 maintenance_skip.go）
- [x] 7.3 更新 `docs/runbooks/release-deploy-and-run.md`（功能开通、邀请码、资格依赖 history、Admin 双入口）
- [x] 7.4 评审自检：无 `g.Redis()` 业务直连、无跨库、无新 ticker、中文注释、不新建 `*_test.go`、主配置无 cash/history 专属回流

## 8. 验收对照

- [x] 8.1 资格：连续 7 日字段正确；VIP 不影响；同日二次走 Redis；history 失败不伪合格
- [x] 8.2 catalog：须绑机；项含 unlocked；预测含 allowedCount；无 UCG 项
- [x] 8.3 支付共用回调分流正确；VIP 购买后功能表无新行
- [x] 8.4 邀请码：单功能兑、同码逐功能、不可自用、失败不绑定、成功绑定一家、换家拒绝、明细可查
- [x] 8.5 Hub 两页可完成功能/SKU/码 CRUD 与兑换追踪，并反映到 App 读接口
