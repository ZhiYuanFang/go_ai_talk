## ADDED Requirements

### Requirement: Clinic Go path SHALL NOT cache or sync conversation history

voice-service Clinic WS MUST NOT 在 Redis 或其他 Go 进程内存储多轮对话 turns，MUST NOT 向客户端下发用于恢复 UI 历史的会话同步帧。对话 UI 历史由客户端本地负责；agent 多轮上下文由 Python 智能服务负责。Go Clinic 路径在成功 `auth_ok` 后 MUST 仅处理 `question`/`cancel` 与流式下行（`thinking_delta`/`answer_delta`/`answer_done`/`turn_cancelled`/`error`），并对 Python 调用仅透传提问所需字段（至少含 `question`、`device_no`、模型配置）。

#### Scenario: auth_ok 后无会话同步帧

- **WHEN** 客户端完成 `auth` 并收到 `auth_ok`
- **THEN** 服务端 MUST NOT 下发 `type=session_sync`
- **AND** 客户端 MAY 立即发送 `question`

#### Scenario: answer_done 不写 Go 会话缓存

- **WHEN** turn 以 `answer_done` 成功结束
- **THEN** voice-service MUST NOT 写入 `voice:clinic:session:{wxId}` 或等价 Go 侧对话缓存键

### Requirement: Clinic Go path SHALL NOT build feeding summary or baby profile for LLM

voice-service 在处理 Clinic `question` 时 MUST NOT 经 `DeviceHistory` 聚合近 7 天喂养摘要，MUST NOT 读写 `voice:clinic:summary:*`，MUST NOT 经 `DeviceProfile` 拉取宝宝画像并注入 Go 侧 LLM system context。喂养上下文与宝宝画像若需要，MUST 由 Python Clinic/companion 路径自行获取。

#### Scenario: question 不因摘要失败而阻断

- **WHEN** history 服务不可用
- **THEN** Clinic `question` 路径 MUST NOT 仅因「喂养摘要加载失败」返回 error（Go 不再依赖该摘要）

#### Scenario: question 不拉 DeviceProfile

- **WHEN** 客户端发送合法 `question`
- **THEN** voice-service Clinic 路径 MUST NOT 为拼装 LLM prompt 调用 DeviceProfile（本路径）

## MODIFIED Requirements

### Requirement: Clinic WebSocket SHALL 以 wxId 为主键绑定身份

`/voice/clinic/ws` 的连接、限流与额度维度 MUST 以 **`wx.id`（正整数）** 为主键，与 `/voice/chat/ws` 以 `deviceNo` 注册 `VoiceWSManager` 的行为 MUST 显式不同。实现 MUST NOT 使用 `deviceNo` 反查 wxId 作为 clinic 鉴权路径（即 MUST NOT 调用 `ResolveVoiceWxID` 的 deviceNo fallback）。`deviceNo` MUST 与 JWT `device_no` claim 一致，并 MUST 透传给 Python Clinic 请求的 `device_no`；MUST NOT 再用于 Go 侧 history 摘要或宝宝画像拉取。

#### Scenario: 首帧 auth 绑定 wxId

- **WHEN** 客户端握手后发送 `type=auth` 且 JWT 含 `sub=1001` 与有效 `device_no`
- **THEN** 服务端 SHALL 解析 `wxId=1001` 并返回 `auth_ok`
- **AND** 后续限流/额度 MUST 使用 wxId=1001

#### Scenario: 未登录拒绝

- **WHEN** 客户端 `auth` 帧 JWT 无效或 `sub≤0`
- **THEN** 服务端 SHALL 返回 `error` code **40301** 且 MUST NOT 进入 `question` 处理

#### Scenario: deviceNo 与 JWT 不一致

- **WHEN** `auth` 帧 `deviceNo` 与 JWT `device_no` claim 不一致
- **THEN** 服务端 SHALL 返回参数错误 `error` 帧且 MUST NOT 返回 `auth_ok`

#### Scenario: 与 voice ball 连接互不踢线

- **WHEN** 同一 `deviceNo` 已建立 `/voice/chat/ws` 且同一 `wxId` 另建 `/voice/clinic/ws`
- **THEN** 两条连接 SHALL 同时保持，且 clinic 连接 MUST NOT 触发 `VoiceWSManager` 替换逻辑

### Requirement: Clinic WebSocket SHALL 使用规定的帧协议

握手成功后，客户端首帧 MUST 为 `type=auth`（含 `accessToken` 与 `deviceNo`）。`auth_ok` 之后，客户端上行 MUST 支持 `type=question` 帧（含非空 `text` 与非空 **`turnId`** UUID）与 **`type=cancel`** 帧（含非空 **`turnId`**）。服务端下行 MUST 支持 `auth_ok`、`thinking_delta`、`answer_delta`、`answer_done`、**`turn_cancelled`**、`error` 六种 `type`，MUST NOT 下发 `session_sync`。流式下行帧（`thinking_delta`、`answer_delta`、`answer_done`）**MUST** 含与当前 turn 一致的 **`turnId`**。`turn_cancelled` MUST 含 **`turnId`** 与 **`reason`**，取值为 **`superseded`**、**`cancelled`** 或 **`disconnected`** 之一。`error` 帧 MUST 含 numeric `code` 与 `message` 字符串。

#### Scenario: 流式 thinking 与 answer 携带 turnId

- **WHEN** 客户端发送 `question` 且 `turnId=uuid-A` 且 LLM 流式返回 reasoning 与 content
- **THEN** 服务端 MUST 推送的 `thinking_delta`、`answer_delta` 与最终 `answer_done` 均 MUST 含 `turnId=uuid-A`

#### Scenario: auth 前拒绝 question

- **WHEN** 客户端未发送 `auth` 或 `auth` 未成功即发送 `question`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: 空问题或缺少 turnId 拒绝

- **WHEN** 客户端发送 `question` 且 `text` 为空或仅空白，或缺少/空 `turnId`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: cancel 中断当前 turn

- **WHEN** 客户端在 turn `uuid-A` 流式进行中发送 `cancel` 且 `turnId=uuid-A`
- **THEN** 服务端 MUST 取消该 turn 的 LLM 上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=cancelled`
- **AND** MUST NOT 下发 `answer_done` 且 MUST NOT consume `clinic_ai`

#### Scenario: 新 question supersede 上一 turn

- **WHEN** turn `uuid-A` 仍在流式进行中且客户端发送新 `question` 且 `turnId=uuid-B`
- **THEN** 服务端 MUST 取消 turn A 的 LLM 上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=superseded`
- **AND** MUST 开始处理 turn B

#### Scenario: WS 断开取消 active turn

- **WHEN** 连接存在 active LLM turn 且 WebSocket 读循环因关闭或错误退出
- **THEN** 服务端 MUST 取消该 turn 的 LLM 上下文
- **AND** MUST NOT consume `clinic_ai`

### Requirement: Clinic SHALL 强制执行 clinic_ai 月度额度（per wxId）

voice-service 在调用 Clinic LLM 前 MUST 使用 auth 已绑定的 `wxId>0`；`wxId≤0` MUST 返回 `error` code **40301**。LLM 调用前 MUST 对 feature `clinic_ai` 以该 wxId 执行 check。若 `allowed=true`，MUST 经 `LoadProfile(LaneClinic)` 调用 LLM；**仅** turn 以 **`answer_done` 成功结束** 时 MUST 以同一 wxId consume。**若 `allowed=false`（`used >= limit`）**，MUST **NOT** 返回 code **40302**；MUST 经 **degraded** 路径调用 LLM，强制 `DefaultSeedProfile(LaneClinic)`（智谱 **`glm-4.1v-thinking-flash`**），且 **`answer_done` 成功时 MUST NOT consume** `clinic_ai`。**cancelled**、**superseded**、**disconnected** 或 LLM 失败而中断的 turn **MUST NOT** consume（含 degraded 路径）。

#### Scenario: 未登录

- **WHEN** wxId 解析为 0 且用户发送 `question`
- **THEN** WS SHALL 返回 40301 且 MUST NOT 调用 LLM

#### Scenario: clinic_ai 额度用尽 degraded 问答

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL **NOT** 返回 40302
- **AND** SHALL 经 degraded 路径流式返回答案
- **AND** `answer_done` 后 MUST NOT consume clinic_ai

#### Scenario: 额度内成功完成扣减

- **WHEN** check 得到 used < limit 且 turn 完整流式结束并下发 `answer_done`
- **THEN** 系统 MUST consume `clinic_ai` 一次

#### Scenario: 用户 cancel 不扣额度

- **WHEN** turn 在流式过程中被 `cancel` 或 `superseded` 结束且未收到 `answer_done`
- **THEN** 系统 MUST NOT 对该 turn 调用 consume `clinic_ai`（含 degraded 路径）

## REMOVED Requirements

### Requirement: Clinic SHALL 注入近 7 天喂养事件聚合摘要

**Reason**: Go 未再向 Python/LLM 注入该摘要；喂养上下文改由 Python 侧负责，避免每轮无效 history 拉取与 Redis 摘要缓存。  
**Migration**: 删除 `clinic_summary.go` / `ensureClinicSummary` / `voice:clinic:summary:*`；客户端与 Python 无需配合改协议。

### Requirement: Clinic 摘要 SHALL 懒刷新

**Reason**: 摘要能力整体移除，懒刷新不再适用。  
**Migration**: 删除摘要 watermark 对比与 Redis refill 逻辑。

### Requirement: Clinic 会话 SHALL 使用固定 12 小时 TTL（wxId 键）

**Reason**: 对话历史不再由 Go Redis 保存；换机/清本地不要求服务端补历史。  
**Migration**: 删除 `voice:clinic:session:{wxId}` 读写与 `appendClinicTurn`；多轮由 Python companion session 承担。

### Requirement: Clinic SHALL 在 auth_ok 后下发 session_sync

**Reason**: 产品要求不再下发；UI 历史由 Flutter 本地存储，且现网客户端已忽略该帧。  
**Migration**: `auth_ok` 后直接进入读循环；旧客户端忽略缺失帧即可继续提问。

### Requirement: Clinic SHALL 注入宝宝画像至 LLM system context

**Reason**: `loadClinicBabyProfile` 结果未传入 `ClinicStream`，属死路径；画像由 Python 拉取。  
**Migration**: 删除 `clinic_profile.go` 及 question 路径上的 DeviceProfile 调用。
