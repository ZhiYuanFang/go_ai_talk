## ADDED Requirements

### Requirement: History service SHALL persist optional expect prediction fields

The History domain MUST support optional fields `expectStartTime`, `expectEndTime`, and `expectAmount` (JSON lowerCamelCase). Times MUST be unix seconds stored as nullable integers. `expectAmount` MUST be a nullable floating-point quantity (MUST NOT truncate to integer-only storage). These fields MUST be non-required on add. When present on add, the service MUST persist them. List/get and History WebSocket push payloads MUST include stored values when non-null. In-memory history cache patches and remote history client maps MUST carry the same fields. Twin client: `flutter_ai_talk` / `history-expect-fields-flutter`.

History 域 MUST 支持可选 `expectStartTime` / `expectEndTime` / `expectAmount`。时间 MUST 为可空 unix 秒；`expectAmount` MUST 为可空浮点，MUST NOT 仅存整数。add MUST 非必填；有值 MUST 落库。list/get、WS、缓存 patch、remote map MUST 在非空时回传。客户端见 twin `history-expect-fields-flutter`。

#### Scenario: Add without expect fields succeeds

- **WHEN** 客户端调用历史 event/add 且 body 不含任何 expect*
- **THEN** 服务端 MUST 在其它字段合法时成功落库
- **AND** 回传实体上 expect* MUST 可为空（NULL / 省略）

#### Scenario: Add with expectStartTime persists and lists

- **WHEN** add body 含合法 `expectStartTime`
- **THEN** 服务端 MUST 持久化该值
- **AND** 随后 list/WS MUST 回传同一 `expectStartTime`

#### Scenario: Add with expectEndTime persists and lists

- **WHEN** add body 含合法 `expectEndTime`
- **THEN** 服务端 MUST 持久化该值
- **AND** 随后 list/WS MUST 回传同一 `expectEndTime`

#### Scenario: Add with expectAmount persists as float

- **WHEN** add body 含合法浮点 `expectAmount`（例如 `120.5`）
- **THEN** 服务端 MUST 以浮点语义持久化
- **AND** list/WS MUST 回传同一数值（允许 JSON 数字表示差异，MUST NOT 截断为整数丢失小数）

### Requirement: History update SHALL NOT overwrite stored expect fields

On history event/update (including setting `endTime` to stop an active timing) and on end-latest-style partial updates, the service MUST NOT write `expect_start_time` / `expect_end_time` / `expect_amount` columns in the UPDATE statement. Request body expect* MUST be ignored. Zero values in Go structs MUST NOT be used to clear existing non-null expect* via a full-column UPDATE map.

历史 event/update（含停表写 `endTime`）与 end-latest 类局部更新 MUST NOT 在 UPDATE 中写入三 expect 列；MUST 忽略 body 中的 expect*；MUST NOT 用结构体零值经整包 Update map 清空库中已有非空 expect*。

#### Scenario: Stop timing keeps expectStartTime

- **WHEN** 某记录已有 `expectStartTime`，客户端 update 仅更新 `endTime`
- **THEN** 库中 `expectStartTime` MUST 保持不变
- **AND** 即使 body 误带 expect* MUST NOT 覆盖原值

#### Scenario: Full event update map omits expect columns

- **WHEN** 客户端调用 event/update 修改 `eventName` / `startTime` 等且行上已有非空 expect*
- **THEN** 服务端生成的 UPDATE 数据 MUST NOT 包含 expect* 列
- **AND** 库中三 expect* MUST 保持更新前的值
