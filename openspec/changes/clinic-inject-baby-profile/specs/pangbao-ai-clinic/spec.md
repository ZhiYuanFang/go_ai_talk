## ADDED Requirements

### Requirement: Clinic SHALL 注入宝宝画像至 LLM system context

每次处理 `question` 并调用 Clinic LLM 前，系统 MUST 以 auth 锁定的 `deviceNo` 经 **`DeviceProfile` HTTP 契约**（如 device-service internal profile GET）取得宝宝 **`birthday`**（Unix 秒）与 **`sex`**。voice-service MUST NOT 直连 `user` 表或 `dao.User`。

系统 MUST 将画像格式化为单行 JSON 并注入 LLM **system** context，位于既有「近 7 天喂养事件聚合摘要（JSON）」块**之前**，格式为：

`宝宝信息（JSON）：{"birthday":"<YYYY-MM-DD 或 未设置>","gender":"<男|女>","age_months":<非负整数>}`

其中：

- `birthday`：`Birthday>0` 时 MUST 格式化为本地时区 `YYYY-MM-DD`；未设置时 MUST 为字符串 **`未设置`**。
- `gender`：`sex>0` MUST 为 **`男`**，否则 MUST 为 **`女`**（与语音球 `loadDeviceProfile` 口径一致）。
- `age_months`：当 `birthday` 已设置时 MUST 为服务端计算的整月月龄；未设置或 `birthday=0` 时 MUST 为 **`0`**。

画像 MUST **每轮 question 实时拉取**，MUST NOT 写入 `voice:clinic:summary:*` Redis 缓存。

#### Scenario: 已设置生日与性别时注入完整画像

- **WHEN** `deviceNo` 对应 profile 含 `birthday>0` 且 `sex=1`（男）
- **THEN** 注入 LLM 的 system context SHALL 包含 `宝宝信息（JSON）：` 前缀的单行 JSON
- **AND** JSON SHALL 含 `birthday` 为 `YYYY-MM-DD`、`gender` 为 `男`、`age_months` 为大于 0 的整月值

#### Scenario: 画像位于喂养摘要之前

- **WHEN** 系统组装 Clinic LLM system message
- **THEN** `宝宝信息（JSON）` 块 SHALL 出现在 `近7天喂养事件聚合摘要（JSON）` 块之前

#### Scenario: DeviceProfile 契约失败时降级继续

- **WHEN** `DeviceProfile` HTTP 调用失败或返回错误
- **THEN** 系统 MUST 记录可观测 warning 日志（含 `deviceNo` 与错误）
- **AND** MUST 使用降级画像：`birthday="未设置"`、`gender="女"`、`age_months=0`
- **AND** MUST 继续调用 Clinic LLM（MUST NOT 仅因 profile 失败返回 WS error 帧）

#### Scenario: 出生日期未设置

- **WHEN** profile 存在但 `birthday=0`
- **THEN** 注入 JSON SHALL 含 `"birthday":"未设置"` 且 `"age_months":0`

#### Scenario: 服务边界

- **WHEN** voice-service 实现 clinic 画像读取
- **THEN** MUST 经 `DeviceProfile()` 契约 HTTP 访问 device 域
- **AND** MUST NOT import 或调用 `hello/internal/dao` 中 user 表 DAO
