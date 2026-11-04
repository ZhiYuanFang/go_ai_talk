## Why

Flutter twin `history-expect-fields-flutter` 将在历史 **add** 请求中携带可选预计字段，供后续预测准确度审计。后端必须非必填落库，并在 list/get/WS 回传，且 **update/stop 不得覆盖** 已有 `expect*`，老客户端不传仍兼容。

## What Changes

- History 领域模型 / 表增加可选列（或 JSON 等价）：`expect_start_time` / `expect_end_time` / `expect_amount`（对外 JSON：`expectStartTime` / `expectEndTime` / `expectAmount`）。
- `POST .../event/add`（现网路径）接受上述可选字段并持久化；缺省 null。
- list / filter / WS 推送实体 MUST 回传已存 expect*。
- `POST .../event/update`（含停表改 `endTime`）MUST NOT 用请求体覆盖已有非空 expect*（忽略客户端传来的 expect* 或禁止写这些列）。
- **兼容**：字段非必填；不改既有字段语义。采用现有 v1 路径上的 **增量可选**（`omitempty` / NULL）；不强制新 path 版本（若评审按 AGENTS 字面要求升版，另开 v2——见 design）。
- **非目标**：不实现 Flutter 打戳逻辑；不做准确度 UI；不在服务端强制补算预报。

## Capabilities

### New Capabilities

- `history-expect-fields`：History 预计字段持久化、add 接受、update 冻结、list/WS 回传。

### Modified Capabilities

- （无）

## Impact

- `history-service`：DAO/entity、`DeviceHistoryEventAddReq`（或等价）、update 写路径、list/WS 序列化。
- 可能：DB migration（`HISTORY_DB_LINK` / history 库表）。
- Twin：`flutter_ai_talk` / `history-expect-fields-flutter`。
- gateway-app：若仅透传 JSON，通常无需改路由；须确认反代不剥 body 字段。
