## 1. 存储与实体

- [ ] 1.1 History 表（或等价存储）增加可空 `expect_start_time` / `expect_end_time` / `expect_amount`；注明 `history-service` + `HISTORY_DB_LINK` / 本域 `database.default`
- [ ] 1.2 DAO/entity 与 list/WS 序列化暴露 JSON `expectStartTime` / `expectEndTime` / `expectAmount`（omitempty）

## 2. Add / Update 路径

- [ ] 2.1 event/add 请求结构接受可选三字段并 INSERT；缺省不写/null
- [ ] 2.2 event/update（含停表）UPDATE 语句 MUST NOT 覆盖已有 expect*；忽略 body 中的 expect*

## 3. 联调与验收

- [ ] 3.1 与 twin `flutter_ai_talk` / `history-expect-fields-flutter` 对齐字段名；确认 gateway-app 透传 add body
- [ ] 3.2 手工：无 expect* add 成功；带 expect* add 后 list/WS 可见；update endTime 后 expect* 不变
