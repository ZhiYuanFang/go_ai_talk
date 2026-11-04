## Why

Flutter twin `history-expect-fields-flutter` 将在历史 **add** 请求中携带可选预计字段，供后续预测准确度审计。后端必须非必填落库，并在 list/get/WS 回传，且 **update/stop 不得覆盖** 已有 `expect*`，老客户端不传仍兼容。

## What Changes

- History 领域模型 / 表增加**可空**列：`expect_start_time` / `expect_end_time` / `expect_amount`（JSON：`expectStartTime` / `expectEndTime` / `expectAmount`）。
- 时间列为可空 bigint（unix **秒**）；`expect_amount` 为可空 **浮点**（与预测侧 `float64` / Flutter twin 对齐，禁止截成 int）。
- Go 实体与 API 用指针或等价可空类型区分「未传 / NULL」与数值 0；JSON `omitempty`。
- `POST .../event/add` 接受可选三字段并 INSERT；缺省不写列（NULL）。
- list / filter / get / WS 推送 MUST 在非 NULL 时回传已存 expect*；缓存 patch、list select、remote adapter map MUST 带上字段。
- `event/update`、`end-latest` 等写路径：UPDATE **MUST NOT 包含** expect* 列（禁止把零值写进 map 抹库）；忽略 body 中的 expect*。
- **兼容**：字段非必填。默认现有 v1 path **增量可选**；若评审按 AGENTS 强制升版则另开 v2（见 design）。
- **非目标**：不实现 Flutter 打戳；不做准确度 UI；服务端不强制补算预报。

## Capabilities

### New Capabilities

- `history-expect-fields`：History 预计字段持久化、add 接受、update 冻结、list/WS 回传。

### Modified Capabilities

- （无）

## Impact

- `history-service`：migration、DAO/entity、`historyInsertData` / `historyUpdateData`、AddReq、list/WS、cache patch、remote adapter。
- DB：`HISTORY_DB_LINK` / history 库 `database.default`。
- Twin：`flutter_ai_talk` / `history-expect-fields-flutter`。
- gateway-app：确认反代不剥 add body 未知键。
