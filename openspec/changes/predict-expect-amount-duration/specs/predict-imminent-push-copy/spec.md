## MODIFIED Requirements

### Requirement: 预测临近可见推送正文 MUST 使用昵称与 title 模板

当 `voice-service` 在预测临近叫醒路径上准备向绑定用户发送 `bizType=predict_imminent` 的**可见**系统推送时，通知正文（传入推送契约的 `alert`）MUST 以基句 `{宝宝昵称}要{事件名}了` 开头，其中：

- **事件名** MUST 为该叫醒载荷中的 `title`（trim 后的非空字符串），MUST NOT 由服务端展开事件树选取子事件名，MUST NOT 使用「事件将在约 N 分钟内发生」类时间回落文案作为事件名。
- **宝宝昵称** MUST 在发送前经 device 画像契约按 `deviceNo` 读取；当返回的 `babyName` 为空、或画像读取失败时，昵称 MUST 回落为字面量「宝宝」。
- 当载荷含**合法** expect 时，系统 MUST 在基句之后追加统一「约」后缀，且 MUST NOT 使用「需要」措辞。合法判定与时长算法 MUST 遵循本变更 design D1/D3：
  - `expectKind=duration` 且 `expectDurationSec>0`：追加 `，预计这次约` + 中文时长（由秒换算：不足 1 分钟为「不足1分钟」，否则为 `N小时` / `N分钟` / `N小时M分钟`，分钟分量按小时浮点 round，与 Flutter 今日汇总一致）。
  - `expectKind=amount` 且 `expectAmount>0`：追加 `，预计这次约` + 格式化数量 +（trim 后非空）`expectUnit` +（trim 后非空且截断至 32 字）`expectLeafTitle`；单位可空。
  - `expectKind` 空或未知、或 kind 与正值字段不匹配时，MUST 视为非法，不追加后缀。
- 当 expect 缺失或非法时，`alert` MUST 仅为基句（与扩展前一致）。
- 系统 MUST NOT 将 expect 字段写入推送契约的 `data` map（仅影响 `alert`）。
- 各厂商通知**标题**（如「胖宝」）本需求 MUST NOT 要求修改。

#### Scenario: 有昵称与 title

- **WHEN** 画像 `babyName` 为「小宝」且叫醒载荷 `title` 为「换尿布」，无 expect，且其它推送闸均通过
- **THEN** 发往推送契约的 `alert` MUST 等于「小宝要换尿布了」

#### Scenario: 昵称为空回落

- **WHEN** 画像 `babyName` 为空串（或画像读取失败）且 `title` 为「吃奶」，无 expect，且其它推送闸均通过
- **THEN** `alert` MUST 等于「宝宝要吃奶了」

#### Scenario: 不使用旧前缀

- **WHEN** 校验通过并发送可见预测临近推送
- **THEN** `alert` MUST NOT 以「宝宝提醒：」开头

#### Scenario: 追加 duration 后缀

- **WHEN** 基句条件满足，且载荷 `expectKind=duration`、`expectDurationSec=3900`，其它闸通过
- **THEN** `alert` MUST 等于「{昵称}要{title}了，预计这次约1小时5分钟」

#### Scenario: duration 不足一分钟

- **WHEN** 基句条件满足，且 `expectKind=duration`、`expectDurationSec=59`，其它闸通过
- **THEN** `alert` MUST 以「，预计这次约不足1分钟」结尾

#### Scenario: duration 整点小时

- **WHEN** 基句条件满足，且 `expectKind=duration`、`expectDurationSec=3600`，其它闸通过
- **THEN** `alert` MUST 以「，预计这次约1小时」结尾（无「0分钟」）

#### Scenario: 追加 amount 后缀含叶子名

- **WHEN** 基句条件满足，昵称为「小宝」，`title` 为「喂养」，`expectKind=amount`、`expectAmount=120`、`expectUnit=毫升`、`expectLeafTitle=配方奶`，其它闸通过
- **THEN** `alert` MUST 等于「小宝要喂养了，预计这次约120毫升配方奶」

#### Scenario: amount 无单位仍追加

- **WHEN** 基句条件满足，`expectKind=amount`、`expectAmount=120`、`expectUnit` 为空、无叶子，其它闸通过
- **THEN** `alert` MUST 等于「{昵称}要{title}了，预计这次约120」

#### Scenario: 非法 expect 不追加

- **WHEN** 基句条件满足，但 `expectKind=duration` 且 `expectDurationSec` 缺失或 ≤0
- **THEN** `alert` MUST 仅为基句，MUST NOT 追加预计后缀

#### Scenario: 未知 kind 不追加

- **WHEN** 基句条件满足，但 `expectKind` 为空或非 `duration`/`amount`
- **THEN** `alert` MUST 仅为基句
