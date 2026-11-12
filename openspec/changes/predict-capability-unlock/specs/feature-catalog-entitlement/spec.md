## MODIFIED Requirements

### Requirement: 系统 MUST 提供绑机可读的合成功能目录 API

cash-service MUST 提供 `GET /cash/app/api/feature/catalog`：须登录且 `X-Internal-Device-No` 非空；`device_no` MUST 仅取自该内部头，MUST NOT 采信 query/body 覆盖。响应 MUST 一次返回功能目录项（无分页），**MUST 包含上架与下架（`status=0`）的功能定义项**（硬编码已废止的 `prediction_unlock` MUST 仍排除）。每项 MUST 含稳定 `featureId`、展示标题、解锁方式提示、**`status`（0=下架，1=上架）**，以及按功能 `activation_subject` 解析的开通态：`unlocked`（bool）；人对功能按请求者 `wx_id` 的人权益，对机功能按设备权益。已开通项 MUST 含 `unlockMethod` 与 `expiresAt`（unix 秒；`0` 表示永久）。UCG 入场资格 MUST NOT 作为目录项出现。该接口 MUST NOT 计入 App usage 统计。该路径 MUST NOT 加入 Bearer 匿名白名单。下架功能的可售 `products` MUST 为空或不含启用 SKU；新开通仍由服务端拒绝，展示策略由客户端根据 `status` 自行决定。

#### Scenario: 一次返回完整目录且含开通态与 status

- **WHEN** 已登录且已绑机的客户端请求功能目录
- **THEN** 响应 MUST 包含功能项，每项含 `unlocked` 与 `status`，且 MUST NOT 要求 page/pageSize

#### Scenario: 未绑机拒绝

- **WHEN** 请求缺少有效 `X-Internal-Device-No`
- **THEN** 系统 MUST 拒绝，MUST NOT 返回目录数据

#### Scenario: 下架功能仍返回并带 status=0

- **WHEN** 某功能在 Admin 被停用（status=0）且客户端请求 catalog
- **THEN** 目录 API MUST 仍返回该项，且该项 `status` MUST 为 0；客户端可自行决定如何展示

#### Scenario: UCG 不在目录中

- **WHEN** 客户端请求功能目录
- **THEN** 响应 MUST NOT 包含 UCG 入场资格项或用 `unlocked` 表示喂养资格

## ADDED Requirements

### Requirement: 下架功能 MUST 拒绝新开通

当功能定义 `status≠1` 时，点试用、创建功能订单、邀请兑换开通该功能 MUST 被拒绝。已存在的未过期权益在目录中 MUST 仍可反映 `unlocked=true`（若未过期），业务使用（含 imminent）MUST NOT 仅因下架而失效（停售不停用）。

#### Scenario: 下架后无法建单

- **WHEN** 功能已下架用户尝试创建该功能订单
- **THEN** 系统 MUST 拒绝建单

#### Scenario: 下架前已开通仍显示已解锁

- **WHEN** 用户在下架前已获未过期人权益且功能现为下架
- **THEN** catalog 该项 `status` MUST 为 0 且 `unlocked` MUST 为 true
