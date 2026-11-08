## Context

`appointment_next` 仅存 `next_at`。Flutter 需要下次预约备注持久化与改约默认回填。兄弟仓 `smart-prediction-appointment-accuracy-ux` 依赖本契约；本仓先合。

## Goals / Non-Goals

**Goals:**

- 表增加 `remark`；App GET 返回、PUT 写入；**Internal GET 同样返回**（与 App 读字段对齐）。
- `nextAt=0` 清空约定时 **同时清空** remark（无约定无备注）。
- 长度上限：`remark` ≤512（DTO `max-length:512`）。

**Non-Goals:**

- 不改 history 表 remark；不改 isAppointment；不写 Redis pending；internal 仍只读不写。

## Decisions

1. **列名** `remark` VARCHAR(512)，默认 `''`；ensure 幂等 ADD COLUMN；新建表 DDL 含该列。  
2. **App PUT**：正 `nextAt` upsert `next_at`+`remark`（空串清备注但保留约定）；`nextAt=0` 写 `next_at=0` 且 `remark=''`。  
3. **App GET / Internal GET**：无行 → `nextAt=0, remark=""`；有行原样返回。Internal 与 App 读字段对齐，Python 可忽略 remark。  
4. **兼容**：同发后客户端始终传 remark 字符串；服务端以收到的字符串为准。

## Risks / Trade-offs

- [旧客户端不传 remark] → 同发窗口内可能被写成空；接受同发。  
- [备注过长] → 校验截断或 4xx。

## Migration Plan

1. ensure 加列。  
2. API/DTO/service。  
3. 与 Flutter 联调。  
4. 回滚：忽略 remark 列（或停读）。

## Open Questions

- 无。
