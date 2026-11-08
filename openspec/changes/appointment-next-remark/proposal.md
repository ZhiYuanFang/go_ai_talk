## Why

Flutter twin 预约下次约定 sheet 需要持久化「下次备注」，并在改约时默认带回上一次预约备注。现网 `appointment_next` 仅有 `next_at`，GET/PUT `/device/app/api/appointment/next` 无备注字段，客户端无法真持久化。需在本仓扩展契约，供 `flutter_ai_talk` change `smart-prediction-appointment-accuracy-ux` 同发跟读跟写（本仓先合）。

## What Changes

- **表**：`appointment_next` 增加 `remark`（或等价）列；ensure 幂等；空串表示无备注。
- **API**：`GET` / `PUT` `/device/app/api/appointment/next` 响应/请求增加 `remark`；`nextAt=0` 清空约定时备注语义写明（见 design）。
- **Internal**：`GET /device/internal/api/appointment/next` 同样返回 `remark`（与 App 读字段对齐；无行空串）；仍只读、不写。
- **非目标**：不改 history.remark；不驱动 predict Redis；不改预约识别 / pending 矩阵其它字段。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `appointment-next-at`：App 读写与 internal 读下次约定时 MUST 支持备注字符串。

## Impact

- `api/v1` App + internal appointment DTO、device 域 schema / service、controller。
- 客户端 twin：`flutter_ai_talk` `smart-prediction-appointment-accuracy-ux`。
- Python 预测兜底可读 remark（可忽略）；与 App 字段对齐。
- 迁移：ensure 加列；旧行 remark 默认空。
