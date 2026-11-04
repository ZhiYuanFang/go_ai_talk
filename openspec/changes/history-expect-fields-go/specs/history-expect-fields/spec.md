## ADDED Requirements

### Requirement: History service SHALL persist optional expect prediction fields

The History domain MUST support optional fields `expectStartTime`, `expectEndTime`, and `expectAmount` (JSON lowerCamelCase; unix seconds for times; numeric amount). These fields MUST be non-required on add. When present on add, the service MUST persist them. List/get and History WebSocket push payloads MUST include stored values when non-null. Twin client behavior is specified in `flutter_ai_talk` change `history-expect-fields-flutter`.

History 域 MUST 支持可选字段 `expectStartTime` / `expectEndTime` / `expectAmount`（JSON lowerCamelCase；时间为 unix 秒）。add MUST 非必填；有值 MUST 落库。list/get 与 History WS 推送 MUST 在非空时回传。客户端见 twin `history-expect-fields-flutter`。

#### Scenario: Add without expect fields succeeds

- **WHEN** 客户端调用历史 event/add 且 body 不含任何 expect*
- **THEN** 服务端 MUST 在其它字段合法时成功落库
- **AND** 回传实体上 expect* MUST 可为空

#### Scenario: Add with expectStartTime persists and lists

- **WHEN** add body 含合法 `expectStartTime`
- **THEN** 服务端 MUST 持久化该值
- **AND** 随后 list/WS MUST 回传同一 `expectStartTime`

### Requirement: History update SHALL NOT overwrite stored expect fields

On history event/update (including setting `endTime` to stop an active timing), the service MUST NOT replace existing non-null `expectStartTime` / `expectEndTime` / `expectAmount` with request body values. The service MUST ignore expect* on update requests for this change.

历史 event/update（含停表写 `endTime`）MUST NOT 用请求体覆盖库中已有非空 expect*；本变更 MUST 忽略 update 上的 expect*。

#### Scenario: Stop timing keeps expectStartTime

- **WHEN** 某记录已有 `expectStartTime`，客户端 update 仅更新 `endTime`
- **THEN** 库中 `expectStartTime` MUST 保持不变
- **AND** 即使 body 误带 expect* MUST NOT 覆盖原值
