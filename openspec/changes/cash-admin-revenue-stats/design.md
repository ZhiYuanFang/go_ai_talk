## Context

开通功能管理（`resource/public/cash-feature-admin.html`）顶栏是 VIP、各上架功能、群二维码。付费落在同库的 `vip_order` / `feature_order`：`amount_fen` 为标价（分），`status=paid` 且 `channel` 为 `alipay` 或 `apple_iap` 才是实付；`admin` 手工授金额为 0；Apple 退款把状态改成 `refunded`。管理接口已在 `/cash/admin/api/*` 下由 gateway-app 反代，并注入 Admin 口令。

管理者需要把这些实付和自己记下的成本放在同一页对比。成本（抽成、服务器、模型）不在订单里。

## Goals / Non-Goals

**Goals:**

- 顶栏增加「收益统计」，打开时一次算出只读合计和手填流水。
- 新手填表，按名称 + 金额 + 方向增删改查。
- 功能付费按功能名拆开；VIP 为一笔合计；差额 = VIP + 各功能付费 + 手填进账 − 手填出账。
- 可选时间范围同时过滤订单 `paid_at` 与手填发生时间。

**Non-Goals:**

- 不把订单金额抄进新表，不改订单、权益、支付回调。
- 不自动扣 Apple / 支付宝费率，不对接账单。
- 不加 Redis，不加后台循环任务。
- 不改 App API 字段，不改 `maintenance_skip`，不改网关 Bind。

## Decisions

### 1. 新表只记手填流水

表名 `cash_manual_ledger`，由 `EnsureSchema` `CREATE TABLE IF NOT EXISTS` 创建，与订单同库（`cash-service` / `default` / `CASH_DB_LINK`）。

| 列 | 含义 |
|---|---|
| `id` | 自增主键 |
| `direction` | `in` 进账 / `out` 出账 |
| `name` | 名称，不唯一，最长 128 |
| `amount_fen` | 正整数分 |
| `occurred_at` | 发生时间，unix 秒；缺省为写入时的 now |
| `created_at` / `updated_at` | unix 秒 |

删除为物理删除。不设软删。

备选：把 VIP/功能合计也写入本表。否决，因为退款和后续支付会和快照漂移。

### 2. 付费合计现场求和

一次读取最多三条 SQL：

1. `vip_order`：`status=paid` 且 `channel IN ('alipay','apple_iap')`，`SUM(amount_fen)`。
2. `feature_order` 同样过滤，`LEFT JOIN feature_product`、`LEFT JOIN feature_def`，按 `feature_id` 分组。展示名用功能 `title`，空则用 `feature_id`。对不上商品的订单归入一组，名称用 `product_code`。
3. `cash_manual_ledger` 按 `direction` 分别 `SUM(amount_fen)`。

时间：查询参数 `start`、`end` 为 unix 秒，均可省略。有 `start` 则 `>= start`，有 `end` 则 `< end`。订单用 `paid_at`，手填用 `occurred_at`。两参数都空表示全部。

`paid_at=0` 的历史实付：全部查询时仍计入；指定时间范围时因 `paid_at` 不在窗内而不计入。不回填 `paid_at`。

### 3. Admin API 只用新路径

沿用现有 Admin POST 变更风格与 `requireCashAdmin`：

| 方法 | 路径 | 作用 |
|---|---|---|
| GET | `/cash/admin/api/revenue/summary` | 合计（含按功能名的列表）与差额 |
| GET | `/cash/admin/api/revenue/entries` | 手填列表，按 `occurred_at` 倒序 |
| POST | `/cash/admin/api/revenue/entries` | 新增 |
| POST | `/cash/admin/api/revenue/entries/update` | 按 `id` 改方向、名称、金额、发生时间 |
| POST | `/cash/admin/api/revenue/entries/delete` | 按 `id` 删除 |

`direction` 只接受 `in` / `out`。`name` 去空白后非空。`amount_fen` 须 ≥ 1。非法参数不写库。

列表与合计共用同一套 `start`/`end`。页面加载时两个 GET 即可，不做缓存。

### 4. Hub 只加一块面板

`renderTiles` 在「群二维码」后追加 `{ id: 'revenue', label: '收益统计' }`。`selectPanel('revenue')` 只显示新区块。`?panel=revenue` 可直接打开。

金额输入用分，旁注元，与现有 VIP 价格框一致。页上写明：付费是用户支付的标价，抽成和运营成本请记为出账。功能表每行一个功能名和金额，下面再给出四项合计与差额。

## Risks / Trade-offs

- [标价不是到账净额] → 页上说明；抽成由管理者记出账，不在服务端猜费率。
- [订单表按状态全表求和，没有专用索引] → 管理页低频打开，先不建索引；变慢再加 `(status, channel, paid_at)`。
- [功能商品被删或改名] → `LEFT JOIN`，标题取当前 `feature_def.title`，对不上则显示 `product_code`。历史金额不随改名重算进别的功能，只是展示名变。
- [指定时间范围会漏掉 `paid_at=0` 的旧实付] → 全部（不传时间）仍计入；不迁移旧行。

## Migration Plan

部署 cash-service 后 `EnsureSchema` 建表。无回填。回滚时去掉新路由与页面区块即可，表可留可删，不影响订单。

## Open Questions

无。功能付费按功能名拆开已确认。
