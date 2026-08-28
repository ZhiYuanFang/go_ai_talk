## ADDED Requirements

### Requirement: 系统 MUST 提供绑机可读的合成功能目录 API

cash-service MUST 提供 `GET /cash/app/api/feature/catalog`：须登录且 `X-Internal-Device-No` 非空；`device_no` MUST 仅取自该内部头，MUST NOT 采信 query/body 覆盖。响应 MUST 一次返回全部启用功能项（无分页）。每项 MUST 含稳定 `featureId`、展示标题、解锁方式提示，以及该设备维度的 `unlocked`（bool）。已开通项 MUST 含 `unlockMethod` 与 `expiresAt`（unix 秒；`0` 表示永久）。UCG 入场资格 MUST NOT 作为目录项出现。该接口 MUST NOT 计入 App usage 统计。该路径 MUST NOT 加入 Bearer 匿名白名单。

#### Scenario: 一次返回完整目录且含开通态

- **WHEN** 已登录且已绑机的客户端请求功能目录
- **THEN** 响应 MUST 包含全部启用功能项，每项含 `unlocked`，且 MUST NOT 要求 page/pageSize

#### Scenario: 未绑机拒绝

- **WHEN** 请求缺少有效 `X-Internal-Device-No`
- **THEN** 系统 MUST 拒绝，MUST NOT 返回目录数据

#### Scenario: 停用功能不出现在目录

- **WHEN** 某功能在 Admin 被停用
- **THEN** 目录 API MUST NOT 再返回该项（缓存失效后或 TTL 内一致语义以实现为准）

#### Scenario: UCG 不在目录中

- **WHEN** 客户端请求功能目录
- **THEN** 响应 MUST NOT 包含 UCG 入场资格项或用 `unlocked` 表示喂养资格

### Requirement: 过期权益 MUST NOT 视为已开通

当某权益 `expiresAt > 0` 且已早于当前时间，对应目录项 `unlocked` MUST 为 false（或等价不可用）。

#### Scenario: 过期后 unlocked 为 false

- **WHEN** 设备曾开通某功能但权益已过期
- **THEN** 目录中该项 `unlocked` MUST 为 false

### Requirement: 功能权益权威数据 MUST 落 MySQL 且按 device_no 维度

功能定义与设备权益 MUST 持久化在 `ai_voice_cash`。权益主体 MUST 为 `device_no`。其他服务 MUST NOT 直查该库。VIP 账号权益 MUST 使用独立 VIP 表，MUST NOT 因 VIP 购买写入功能权益行。

#### Scenario: 开通写入本域库

- **WHEN** 支付/邀请码/广告成功开通某功能
- **THEN** cash-service MUST 在 `ai_voice_cash` 写入或更新对应 `device_no` 权益行（或更新 `allowedCount`）

### Requirement: App MUST NOT 依赖独立 entitlements 列表接口作为主读模型

本变更 App 主读模型 MUST 为合成 catalog。系统 MUST NOT 将独立 `GET .../feature/entitlements` 作为客户端必调接口（可省略该 App 路径）。

#### Scenario: 客户端仅调 catalog 即可展示开通态

- **WHEN** 客户端已调用合成 catalog
- **THEN** 响应 MUST 足以按列表展示各功能是否已开通，无需再调独立权益列表
