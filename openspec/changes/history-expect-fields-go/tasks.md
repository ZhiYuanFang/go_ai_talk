## 1. 存储与实体

- [x] 1.1 History 表增加可空 `expect_start_time` / `expect_end_time`（BIGINT 秒）与 `expect_amount`（浮点，非 INT）；注明 `history-service` + `HISTORY_DB_LINK` / `database.default`
- [x] 1.2 DAO/entity 用可空类型（`*int64` / `*float64` 或 Null*）；JSON `expectStartTime` / `expectEndTime` / `expectAmount` + omitempty；list select、`historyRowToEntity`、cache patch、WS payload、remote adapter map 全链路带上

## 2. Add / Update 路径

- [x] 2.1 event/add Req 接受可选三字段；`historyInsertData` 仅在非 nil 时写入列；缺省 NULL
- [x] 2.2 `historyUpdateData` **禁止**包含三 expect 列；event/update 忽略 body expect*；确认 end-latest 仍只改 endTime/remark

## 3. 联调与验收

- [x] 3.1 与 twin `history-expect-fields-flutter` 对齐字段名与 amount 浮点；确认 gateway-app 透传 add body
- [x] 3.2 手工：无 expect* add 成功；带三字段 add 后 list/WS 可见（含小数 amount）；update / 停表后 expect* 不变
