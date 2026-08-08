## ADDED Requirements

### Requirement: care-alert 必须携带有效 wxId

voice-service 上经 gateway 暴露的 care-alert HTTP（`GET /device/api/care-alert/daily`、`DELETE /device/api/care-alert/daily/item`、`POST /device/api/care-alert/feedback`）MUST 要求请求携带有效 `X-Internal-Wx-Id`（解析后 `wxId > 0`）。纯设备会话（`wxId=0` 或缺失头）MUST 被拒绝。鉴权路径 MUST NOT 使用按 `deviceNo` 反查 `wxId` 作为登录替代。

#### Scenario: 缺少或无效 wxId

- **WHEN** 客户端调用任一 care-alert 接口且 `X-Internal-Wx-Id` 缺失、非正整数或为 `0`
- **THEN** 服务 MUST 返回错误且 MUST NOT 读写日缓存生成路径中的 LLM，也 MUST NOT 完成需登录的删除/飞轮成功语义

#### Scenario: 有效 wxId 与 deviceNo

- **WHEN** 请求同时具备 `wxId > 0` 与合法 `deviceNo`
- **THEN** 服务 MAY 继续执行宝宝维度的缓存/忽略/飞轮逻辑；VIP 判定主键 MUST 为该 `wxId`

### Requirement: 按触发者 VIP 选模且与日缓存维度分离

护理留意首次日生成（缓存未命中）时，voice-service MUST 按**当前请求触发者**的 `wxId` 查询账号 VIP：VIP 为真 MUST 选用 DeepSeek；非 VIP MUST 选用 Zhipu。日缓存键 MUST 仍为 `deviceNo + Asia/Shanghai` 自然日；缓存命中 MUST NOT 因读缓存再次查询 VIP 或重跑 LLM。本路径 MUST NOT 扣减 clinic AI 配额。

#### Scenario: VIP 触发者首次生成

- **WHEN** 某 `deviceNo` 当日缓存未命中，且触发者 `wxId` 经 device 契约判定为 VIP
- **THEN** 服务 MUST 以 DeepSeek 配置调用 Python 分析，并将结果写入该 `deviceNo` 当日缓存

#### Scenario: 非 VIP 触发者首次生成

- **WHEN** 某 `deviceNo` 当日缓存未命中，且触发者判定为非 VIP
- **THEN** 服务 MUST 以 Zhipu 配置调用 Python 分析并写入当日缓存

#### Scenario: 他看护命中缓存

- **WHEN** 同 `deviceNo` 同日已有缓存，另一 `wxId`（无论是否 VIP）再次 GET
- **THEN** 服务 MUST 直接返回缓存列表，MUST NOT 按后者 VIP 重新选模或重跑 LLM

### Requirement: VIP 查询失败降级为非 VIP

voice-service 在 care-alert 生成路径查询账号 VIP 时，若 device 契约超时、不可达、或返回不可用错误，MUST 将本次生成视为非 VIP（Zhipu），MUST NOT 因此使整个 daily 生成失败（设备校验与 Python 分析等其它错误仍按原语义失败）。实现 MUST 记录可观测 Warning。

#### Scenario: device VIP 接口失败

- **WHEN** 缓存未命中且查询 `vip-by-wx-id`（或等价契约）失败
- **THEN** 服务 MUST 选用 Zhipu 继续生成（若其它步骤成功），且 MUST 打出 Warning 级日志表明 VIP 降级
