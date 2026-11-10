## MODIFIED Requirements

### Requirement: History update SHALL NOT overwrite stored expect fields

On history event/update (including setting `endTime` to stop an active timing) and on voice/intent `event/batch` update/end operations that use the general history update column set, the service MUST NOT replace existing non-null `expectStartTime` / `expectEndTime` / `expectAmount` with request body values. The service MUST ignore expect* on those update requests. This freeze MUST NOT apply to the dedicated expect-batch endpoint defined in this change, which MAY write expect* columns.

历史 event/update（含停表）以及走通用 update 列集的 `event/batch` update/end MUST NOT 用请求体覆盖库中已有非空 expect*；MUST 忽略其上的 expect*。本冻结 MUST NOT 作用于本变更定义的 expect-batch 专用接口（该接口 MAY 写 expect* 列）。

#### Scenario: Stop timing keeps expectStartTime

- **WHEN** 某记录已有 `expectStartTime`，客户端 update 仅更新 `endTime`
- **THEN** 库中 `expectStartTime` MUST 保持不变
- **AND** 即使 body 误带 expect* MUST NOT 覆盖原值

#### Scenario: Batch CRUD update still freezes expect

- **WHEN** 客户端经 `event/batch` op=update 更新备注或时间且 body 含 expect*
- **THEN** 服务端 MUST NOT 用该 expect* 覆盖库中已有值

## ADDED Requirements

### Requirement: History service SHALL expose expect-batch patch endpoint

The History service MUST provide an authenticated device API at `POST /device/history/api/event/expect-batch` that accepts `deviceNo` and a non-empty list of items each with history `id` and optional `expectStartTime` / `expectEndTime` / `expectAmount`. An empty `items` array MUST result in HTTP 400. More than 200 items MUST result in HTTP 400. Present keys MUST update the corresponding stored expect columns for rows that belong to that `deviceNo`; omitted keys MUST leave existing column values unchanged. The service MUST NOT treat JSON null as a supported clear-to-NULL operation in v1. The service MUST NOT require clients to resend eventId/startTime/remark to patch expect*. Items that fail (foreign id, missing row, invalid id, or write error) MUST be skipped without rolling back other items' successful updates. The response MUST be `{ "ok": true }` when at least one item was successfully updated, and `{ "ok": false }` when none were; the response MUST NOT include per-item results or per-item error details. For a well-formed batch (non-empty items within limit), the HTTP status MUST be 200 regardless of `ok`. Twin client: `flutter_ai_talk` change `history-expect-anchor-chain-scheduler`.

History 服务 MUST 提供鉴权 device 接口 `POST /device/history/api/event/expect-batch`，接受 `deviceNo` 与非空 `{id, expectStartTime?, expectEndTime?, expectAmount?}` 列表。空 `items` MUST 返回 HTTP 400；超过 200 条 MUST 返回 HTTP 400。出现的键 MUST 更新归属该 deviceNo 的对应列；省略的键 MUST 保持原值。v1 MUST NOT 将 JSON null 当作清列语义。MUST NOT 要求为打 expect 补丁而重传 eventId/startTime/remark。失败条目（外机、不存在、无效 id、写错）MUST 跳过且 MUST NOT 回滚其它已成功更新。只要 ≥1 条成功更新响应 MUST 为 `{ "ok": true }`；全部未写上 MUST 为 `{ "ok": false }`；MUST NOT 返回 per-item results/逐条错误。合法 batch（非空且未超限）无论 `ok` 真假 HTTP 均 MUST 为 200。客户端 twin：`history-expect-anchor-chain-scheduler`。

#### Scenario: Batch patches expectStartTime for owned ids

- **WHEN** 合法 Bearer 调用 expect-batch，items 含本 device 下两行 id 及新的 `expectStartTime`
- **THEN** 服务端 MUST 持久化这两行的新 `expectStartTime`
- **AND** 未出现在请求中的 expectEndTime/expectAmount MUST 保持原值（若库中已有）
- **AND** 响应 MUST 为 `{ "ok": true }`

#### Scenario: Empty items rejected

- **WHEN** 合法 Bearer 调用 expect-batch 且 `items` 为空数组
- **THEN** 服务端 MUST 返回 HTTP 400
- **AND** MUST NOT 修改任何行

#### Scenario: Foreign id does not block other patches

- **WHEN** items 中混有本机合法 id 与不属于该 `deviceNo` 的 id
- **THEN** 服务端 MUST 持久化合法 id 的 expect 变更
- **AND** MUST NOT 因外机 id 回滚已成功行
- **AND** 因至少一条成功，响应 MUST 为 `{ "ok": true }`
- **AND** 响应 MUST NOT 包含逐条错误明细

#### Scenario: All items fail yields ok false

- **WHEN** 合法 batch 中每条均无法更新（例如 id 全不存在）
- **THEN** 服务端 MUST 返回 HTTP 200
- **AND** 响应 MUST 为 `{ "ok": false }`
- **AND** MUST NOT 修改任何行

#### Scenario: Omitting expectAmount leaves amount unchanged

- **WHEN** 某行已有 `expectAmount`，batch item 仅带 `expectStartTime`
- **THEN** 库中 `expectAmount` MUST 不变

### Requirement: Expect-batch success SHALL notify with full expect fields

After a successful expect-batch item write, the service MUST reload the history row (or otherwise assemble a complete entity including stored expect*) before updating any in-process cache and before publishing the History WebSocket `update` event for that item. Failed/skipped items MUST NOT emit a stripping update. The service MUST NOT publish an update payload that strips previously stored expect* solely because the patch writer omitted reconstructing them.

expect-batch 单条写成功后，服务端 MUST 在更新缓存与推送 History WS `update` 前 reload（或组装）含已存 expect* 的完整行。跳过/失败的条目 MUST NOT 推送剥字段的 update。MUST NOT 因补丁组装不完整而推送剥掉 expect* 的 update。

#### Scenario: WS update after expect-batch includes expectStartTime

- **WHEN** expect-batch 成功将某行 `expectStartTime` 从空改为有值
- **THEN** 随后 WS update（或等价推送）MUST 携带该 `expectStartTime`

### Requirement: Expect-batch App API SHALL NOT count toward usage stats

The gateway-app usage stats path MUST treat `POST /device/history/api/event/expect-batch` as a maintenance/excluded API (exact method+path in `usagestats` maintenance skip), per owner confirmation for this change. Successful or failed calls MUST NOT increment App API usage counters for that apiKey.

gateway-app 使用统计 MUST 将 `POST /device/history/api/event/expect-batch` 记为维护/排除接口（`usagestats` maintenance 精确 method+path），按本变更负责人确认。成功或失败调用 MUST NOT 计入该 apiKey 的 App API usage。

#### Scenario: Expect-batch excluded from usage

- **WHEN** 客户端经 gateway-app 调用 `POST /device/history/api/event/expect-batch`
- **THEN** 该请求 MUST NOT 计入 App API 使用统计
