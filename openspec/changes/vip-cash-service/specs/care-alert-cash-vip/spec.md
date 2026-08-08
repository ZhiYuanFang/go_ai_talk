## ADDED Requirements

### Requirement: care-alert 经 cash-service 判定触发者 VIP

voice-service 在护理留意日缓存未命中、需要选模时，MUST 使用当前请求触发者的 `wxId`，经 HTTP 调用 cash-service 内部 VIP 接口判定是否 VIP：VIP MUST 选用 DeepSeek；非 VIP MUST 选用 Zhipu。MUST NOT 查询 `wx.is_vip`，MUST NOT 调用 device-service 的 VIP 接口，MUST NOT 使用 `deviceNo` 反查作为 VIP 或登录依据。日缓存键 MUST 仍为 `deviceNo + Asia/Shanghai` 自然日；缓存命中 MUST NOT 重跑 LLM。本路径 MUST NOT 扣减 clinic AI 配额。

#### Scenario: VIP 触发者首次生成

- **WHEN** 当日缓存未命中且 cash 返回该触发者 `isVip=true`
- **THEN** 服务 MUST 以 DeepSeek 配置调用 Python 分析并写入宝宝日缓存

#### Scenario: 非 VIP 触发者首次生成

- **WHEN** 当日缓存未命中且 cash 返回 `isVip=false`
- **THEN** 服务 MUST 以 Zhipu 配置调用 Python 分析并写入宝宝日缓存

#### Scenario: cash VIP 查询失败降级

- **WHEN** 调用 cash 内部 VIP 接口超时或失败
- **THEN** 服务 MUST 按非 VIP（Zhipu）继续生成（其它步骤成功时），MUST 打 Warning，MUST NOT 因 VIP 查询失败单独使整个 daily 失败

### Requirement: care-alert 必须携带有效 wxId

`GET/DELETE/POST` 之 `/device/api/care-alert/*` MUST 要求 `X-Internal-Wx-Id` 解析后 `wxId>0`；纯设备会话 MUST 被拒绝。

#### Scenario: 缺少 wxId

- **WHEN** care-alert 请求缺少有效 `X-Internal-Wx-Id`
- **THEN** 服务 MUST 返回错误且 MUST NOT 完成需登录的生成/删除/飞轮成功语义

### Requirement: 拆除 device 域 wx.is_vip 路径

本变更 MUST 移除（或停止注册）下列由 `care-alert-vip-by-wx` 引入的路径，使 VIP 真相源唯一落在 cash-service：`wx` 表 `is_vip` 模型字段与 DDL 执行指引、`GET /device/app/api/user/internal/vip-by-wx-id`、voice 经 device 的 `RemoteIsVipByWxID`。`hack/ddl_wx_is_vip.sql` MUST NOT 作为发布必执行项。

#### Scenario: device 不再提供 VIP 内部接口

- **WHEN** 客户端或 voice 请求已删除的 device `vip-by-wx-id` 路径
- **THEN** 该接口 MUST 不可用（404 或未注册），调用方 MUST 改用 cash internal VIP 接口

#### Scenario: 发布文档不再要求执行 is_vip DDL

- **WHEN** 按本变更更新后的 runbook 部署 VIP 能力
- **THEN** 文档 MUST 指引部署 `cash-service` / `ai_voice_cash`，MUST NOT 要求执行 `ddl_wx_is_vip.sql` 作为 VIP 前置
