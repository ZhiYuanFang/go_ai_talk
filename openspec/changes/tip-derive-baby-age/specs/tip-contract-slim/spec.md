## ADDED Requirements

### Requirement: Go App Tip API 请求体去掉月龄与当前时间

`POST /device/tip/generate` 的请求体（`DeviceTipGenerateReq` 或接线期等价结构）SHALL NOT 再包含 `babyAgeMonths`、`currentTime` 字段。相关 dc/注释 MUST NOT 再声称「0 则服务端按生日推算」或「0 则使用服务端当前时间」。

#### Scenario: 精简后的 App 请求体

- **WHEN** Flutter 或其它 App 客户端调用 `POST /device/tip/generate`
- **THEN** 请求 JSON SHALL 可仅含设备与事件相关字段（如 `deviceNo`、`eventId`、`eventName`）及既有其它非月龄/时间必填项
- **AND** 服务端 SHALL NOT 要求 `babyAgeMonths` 或 `currentTime`

#### Scenario: 假注释消除

- **WHEN** 评审 `api/v1/device_tip_http.go`（及控制器/服务注释）
- **THEN** SHALL NOT 存在「月龄为 0 由服务端推算」且实际未实现的误导描述

### Requirement: Go TipStream 与 PythonAIClient 去掉对应参数

`VoiceService.TipStream` 与 `PythonAIClient.TipStreamRequest` SHALL 删除 `babyAgeMonths`/`BabyAgeMonths`、`currentTime`/`CurrentTime`（及 JSON `baby_age_months`、`current_time`）。Go MUST NOT 再为 tip 补全 `currentTime` 后传给 Python。

#### Scenario: TipStream 签名瘦身

- **WHEN** 控制器调用 `TipStream` 发起小贴士流
- **THEN** 调用签名 SHALL 不再接收月龄与当前时间参数
- **AND** 发往 Python 的 JSON body SHALL 不含 `baby_age_months`、`current_time`

### Requirement: Python TipRequest 删除或废弃入站月龄与时间字段

`TipRequest` SHALL 删除 `baby_age_months`、`current_time` 字段，或将其标为废弃且非必填并忽略；权威月龄与时间上下文 MUST 来自服务端派生（见 `tip-python-derived-context`）。

#### Scenario: Tip 入站无月龄时间字段可成功

- **WHEN** Go 以 snake_case 发送 tip 请求且不含 `baby_age_months`、`current_time`
- **THEN** Python SHALL 接受该请求（其它必填合法时）
- **AND** SHALL 使用派生月龄与派生当前时间生成提示词

### Requirement: Flutter tip 调用去掉月龄与 currentTime

Flutter `tip_repository` / `tip_provider` / `home_screen` SHALL 不再计算或传递 `babyAgeMonths`，且请求体 SHALL NOT 包含 `currentTime`。

#### Scenario: home 触发 tip 不再传月龄

- **WHEN** `home_screen` 因历史 create 事件触发 tip 流式生成
- **THEN** 调用链 SHALL NOT 再传入 `babyAgeMonths`
- **AND** POST body SHALL NOT 包含 `babyAgeMonths` 或 `currentTime`
