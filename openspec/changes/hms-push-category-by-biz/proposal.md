## Why

华为 Push 对云通知按消息分类管控：未声明或错误的 `category` 易被归入资讯营销，送达与展示受限。现有 `predict_imminent`（工作提醒）与 `ucg_alert`（社区/私信类）已有稳定 `bizType`，但 HMS 下发载荷未填 `android.notification.category`，无法按自分类策略走服务与通讯通道。

## What Changes

- HMS 发送路径在构建 `android.notification` 时，按 `bizType` 写入华为官方枚举 `category`：
  - `predict_imminent` → `WORK`
  - `ucg_alert` → `IM`
- `ucg_silent_badge` **不传** `category`（静默角标，不走可见分类）。
- 未知或空 `bizType` **不写** `category`（避免误标）。
- 不改 APNs/MiPush、不改跨服务推送契约与 `bizType` 常量集合。

## Capabilities

### New Capabilities

- `push-hms-message-category`：push-service 经 HMS 下发时，`android.notification.category` 与 `bizType` 的映射及省略规则。

### Modified Capabilities

- （无）本变更为新增 HMS 分类行为，不修改既有基线 capability 的已归档需求条文。

## Impact

- **代码**：`internal/services/push/push_hms.go`（`androidNotif` 构建）
- **进程**：仅 `push-service`
- **依赖**：华为侧需已具备对应自分类权限；无权限时行为以厂商响应为准（本变更不新增申请流程）
- **不涉及**：DB、Redis、gateway、voice/ucg 调用方、其它厂商通道
