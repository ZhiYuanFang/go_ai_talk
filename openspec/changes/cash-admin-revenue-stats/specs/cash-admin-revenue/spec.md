## ADDED Requirements

### Requirement: 管理端 MUST 提供手填进账与出账

cash-service MUST 在与 `vip_order` 同库的新表中保存手填流水。每条 MUST 包含方向（`in` 进账或 `out` 出账）、非空名称、大于等于 1 的金额（分）和发生时间（unix 秒）。名称 MUST NOT 要求唯一。Admin MUST 能新增、按 id 修改上述字段、按 id 物理删除、并按发生时间倒序列出。方向、空名称或金额小于 1 MUST 拒绝且 MUST NOT 写库。未传发生时间时 MUST 使用写入时刻。

#### Scenario: 新增一笔出账

- **WHEN** 管理员提交方向 `out`、名称「模型费用」、金额 5000 分
- **THEN** 系统 MUST 写入该行，且列表中能按名称与金额读回

#### Scenario: 金额为 0

- **WHEN** 管理员提交金额 0
- **THEN** 系统 MUST 拒绝，且 MUST NOT 插入行

### Requirement: 收益合计 MUST 只统计实付并按功能名拆开

管理端收益合计 MUST 只把 `status=paid` 且 `channel` 为 `alipay` 或 `apple_iap` 的订单 `amount_fen` 算作付费。`channel=admin`、`refunded` 以及其他非 paid 状态 MUST NOT 计入。VIP 付费 MUST 为一个合计。功能付费 MUST 按功能拆开：展示名为功能标题，标题为空时 MUST 用功能编号；对不上功能商品的订单 MUST 单独成组并以商品编码为名称。合计 MUST 在读取时对订单求和，MUST NOT 把订单金额写入手填表。

#### Scenario: 功能付费按名称列出

- **WHEN** 两个不同功能各有一笔已支付的支付宝或 Apple 订单
- **THEN** 合计 MUST 返回两行，每行含该功能名称与对应金额之和

#### Scenario: 手工授与退款不计入

- **WHEN** 存在 `channel=admin` 的 paid 订单，以及 `status=refunded` 的订单
- **THEN** 这两类订单的金额 MUST NOT 出现在 VIP 或功能付费合计中

### Requirement: 差额 MUST 用手填与付费一起算出

在同一时间范围内，系统 MUST 计算差额 = VIP 付费 + 各功能付费之和 + 手填进账之和 − 手填出账之和。查询参数 `start`、`end`（unix 秒）均可省略。省略两者时 MUST 统计全部。给出 `start` 时订单 MUST 满足 `paid_at >= start`、手填 MUST 满足发生时间 `>= start`。给出 `end` 时对应时间 MUST `< end`。订单过滤与手填过滤 MUST 使用同一组参数。

#### Scenario: 对比成本与收益

- **WHEN** VIP 实付 1900 分、某功能实付 600 分、手填进账 0、手填出账 5000 分，且未传时间范围
- **THEN** 差额 MUST 为 1900 + 600 − 5000

#### Scenario: 时间范围同时作用于两边

- **WHEN** 管理员传入 `start` 与 `end`
- **THEN** 付费合计 MUST 只含该窗口内 `paid_at` 的实付，手填合计 MUST 只含该窗口内发生的流水

### Requirement: 开通功能管理 MUST 提供收益统计面板

`cash-feature-admin.html` 顶栏 MUST 在现有入口之外提供「收益统计」。选中后 MUST 只显示该面板，并展示 VIP 合计、按功能名的付费列表、手填进账合计、手填出账合计、差额，以及手填流水的增删改查。页面 MUST 说明付费金额是用户支付的标价，渠道抽成与运营成本须记为出账。金额输入 MUST 以分为单位。

#### Scenario: 打开收益统计

- **WHEN** 管理员在开通功能管理选中「收益统计」
- **THEN** 页面 MUST 展示上述合计与手填列表，且 MUST NOT 改动订单或权益数据
