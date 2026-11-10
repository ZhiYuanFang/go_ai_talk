## ADDED Requirements

### Requirement: Imminent Live Activity title MUST be composed as nick要事件了 in history Refresh

When the unified Live Activity Refresh projects a `kind=imminent` item into ContentState `items`, the `title` field MUST be a finished Chinese sentence of the form `{nick}要{eventName}了`, where `eventName` is the trimmed event display name from Redis imminent pending `title`, and `nick` is the device baby display name resolved via the existing history→device profile contract (e.g. `GetBirthday` / profile). When the baby name is empty or profile lookup fails, `nick` MUST fall back to `宝宝`. The composition MUST occur only on the history Live Activity projection path (Refresh / `buildLiveActivityItems` or equivalent). The system MUST NOT require ordinary `predict_imminent` visible-push code to write this sentence into Redis or into Live Activity ContentState. Redis pending MAY continue to store the bare event name as reported by the client. When the pending event name is empty after trim, the projected item MUST omit `title` (client empty fallback MAY apply) and MUST NOT invent a nick-only sentence.

统一 Live Activity Refresh 投影 `kind=imminent` 时，ContentState `title` MUST 为 `{nick}要{事件名}了`；nick 经 history→device 画像取得，空或失败回落「宝宝」；拼装 MUST 仅在 history LA 投影路径完成；MUST NOT 依赖普通可见推送路径写入 Redis/LA。pending 可仍存裸事件名；事件名为空时 MUST 省略 `title`。

#### Scenario: Composed title with baby name

- **WHEN** Refresh 投影 imminent 项且 pending.title=`吃奶`、设备宝宝名为`豆豆`
- **THEN** ContentState 该项 `title` MUST 为`豆豆要吃奶了`

#### Scenario: Fallback nick when profile empty

- **WHEN** Refresh 投影 imminent 项且 pending.title=`睡觉`、宝宝名为空或画像失败
- **THEN** ContentState 该项 `title` MUST 为`宝宝要睡觉了`

#### Scenario: Pending sync Refresh without ordinary push still composes

- **WHEN** App PUT imminent pending 成功触发 Refresh，且尚无对应 MQ 可见推送
- **THEN** 投影出的 imminent `title` MUST 仍为成品句（MUST NOT 因未走普通推送而只透传裸事件名）

#### Scenario: Empty event name omits title

- **WHEN** 进窗 pending 项 eventId/nextAt 合法但 title trim 后为空
- **THEN** 投影项 MUST 省略 `title` 键，MUST NOT 写出仅含 nick 的句子

### Requirement: Imminent Live Activity title MUST NOT embed expect amount or duration copy

When projecting imminent items, structured expect fields (`expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit`) MUST continue to be written as separate ContentState keys when usable (existing apply-pending-expect rules). The composed `title` MUST NOT append ordinary-push expect suffixes (including「，预计这次约…」or leaf titles). Clients MAY render short expect copy in a trailing lock-screen slot from those structured fields.

投影 imminent 时 expect MUST 仍为独立 ContentState 键；成品 `title` MUST NOT 追加普通推送的预计后缀或 leaf；客户端 MAY 用结构化字段渲染右侧短文案。

#### Scenario: Duration expect stays out of title

- **WHEN** pending 项 title=`吃奶`、nick=`豆豆`、expectKind=`duration`、expectDurationSec=`1800`
- **THEN** `title` MUST 为`豆豆要吃奶了`且 MUST NOT 含`预计`
- **AND** ContentState MUST 仍含 `expectKind=duration` 与 `expectDurationSec=1800`

#### Scenario: Amount expect stays out of title

- **WHEN** pending 项 title=`吃奶`、nick=`豆豆`、expectKind=`amount`、expectAmount=`120`、expectUnit=`ml`
- **THEN** `title` MUST 为`豆豆要吃奶了`
- **AND** ContentState MUST 仍含 amount 相关 expect 键
