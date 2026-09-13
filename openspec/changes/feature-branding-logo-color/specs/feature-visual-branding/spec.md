## ADDED Requirements

### Requirement: feature_def MUST 持久化 logo 与 color

`ai_voice_cash.feature_def` MUST 包含 `logo` 与 `color` 列。`logo` MUST 存储 OSS objectKey（建议前缀 `feature/`），MUST NOT 在库内持久化完整 CDN URL 作为权威值。`color` MUST 为 `#RGB` 或 `#RRGGBB`（大小写不敏感）或空串。cash-service 以外进程 MUST NOT 直查该表。

#### Scenario: 写入 objectKey 与 hex

- **WHEN** 管理员保存某功能的 logo 与主色
- **THEN** 系统 MUST 将 objectKey 与校验通过的 color 持久化到该 `feature_id` 行

#### Scenario: 非法 color 拒绝

- **WHEN** 提交的 color 不是空串且不符合 `#RGB`/`#RRGGBB`
- **THEN** 系统 MUST 拒绝保存并返回错误

### Requirement: 系统 MUST 为三个内置功能种子默认 color

EnsureSchema/种子 MUST 为下列功能写入默认 `color`（logo 默认为空串）：

| feature_id | color |
|------------|-------|
| `prediction_unlock` | `#3B82F6` |
| `care_alert_smart_remind` | `#0D9488` |
| `growth_trajectory_predict` | `#EA580C` |

当行已存在且 `color` 非空时，种子 MUST NOT 覆盖该 color。

#### Scenario: 新库种子带默认色

- **WHEN** 全新库执行功能定义种子
- **THEN** 上述三个 `feature_id` 的 `color` MUST 分别为表中默认值且 `logo` 为空

#### Scenario: 运维已改色不被种子覆盖

- **WHEN** 某功能 `color` 已被运维改为非空自定义值后再次跑种子
- **THEN** 该行 `color` MUST 保持运维值

### Requirement: App 功能目录 MUST 下发 logo 与 color

`GET /cash/app/api/feature/catalog` 每项 MUST 包含 `logo` 与 `color`。HTTP 边界上 `logo` MUST 为可访问的 CDN 绝对 URL；若库内 objectKey 为空，则 `logo` MUST 为空串。`color` MUST 原样下发库内值（种子后内置功能通常非空）。该扩展属于 v1 结构变更；本能力不要求保留无字段旧响应。

#### Scenario: 已配置视觉的目录项

- **WHEN** 某启用功能已配置 logo objectKey 与 color，客户端请求 catalog
- **THEN** 该项 MUST 含非空 CDN `logo` 与对应 `color`，且仍含既有 `unlocked` 等开通态字段

#### Scenario: 未上传 logo

- **WHEN** 某功能 `logo` 库内为空
- **THEN** catalog 该项 `logo` MUST 为空串，`color` 仍 MUST 返回（含种子默认色）

### Requirement: Admin 功能定义 API MUST 读写 logo 与 color

功能定义列表与更新 API MUST 支持 `logo` 与 `color`。列表/读出时 `logo` MUST 映射为 CDN URL（与 App 一致）。更新时客户端可提交 objectKey（上传接口返回值）与 color；未更换 logo 时 MUST 允许保留原 objectKey。

#### Scenario: 更新主色

- **WHEN** 管理员仅修改 color 并保存
- **THEN** 系统 MUST 更新 color，且 MUST NOT 清空已有 logo

### Requirement: 系统 MUST 支持功能 logo 上传至 OSS

系统 MUST 提供 Admin 可调用的功能 logo 上传能力（建议 `POST /cash/admin/api/feature/defs/logo` 或等价），经 ucg 契约写入 OSS，返回 objectKey（及可选 CDN URL）。objectKey MUST 使用 `feature/` 前缀（或与实现约定的等价 feature 命名空间）。实现 MUST 经 `clients/ucg`（或 platform 封装），MUST NOT 从 cash 业务包 import device 服务实现。

#### Scenario: 上传成功返回 objectKey

- **WHEN** 管理员上传合法 png/jpg/jpeg/webp logo
- **THEN** 系统 MUST 返回可写入 `feature_def.logo` 的 objectKey

#### Scenario: 超大或非法类型拒绝

- **WHEN** 文件超过大小上限或扩展名不支持
- **THEN** 系统 MUST 拒绝上传且 MUST NOT 写入 `feature_def`
