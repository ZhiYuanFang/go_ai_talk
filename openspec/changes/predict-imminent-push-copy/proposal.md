## Why

预测临近离线推送的可见正文目前为「宝宝提醒：」+ title（或「事件将在约 N 分钟内发生」回落），读感像系统告警，且未带宝宝昵称。产品希望正文改为口语化的「{昵称}要{事件名}了」，并在缺少事件名时宁可不推，避免尴尬占位文案。

## What Changes

- 预测临近（`bizType=predict_imminent`）可见通知**正文**改为模板：`{宝宝昵称}要{事件名}了`。
- **事件名**取客户端同步 pending 时携带的 `title`（父事件展示名）；不解析事件树叶子，不猜「拉屎/尿尿」。
- **宝宝昵称**在推送前经 device 画像契约读取；空串或画像失败时回落为「宝宝」。
- **`title` 为空（trim 后）时 MUST NOT 发送可见推送**（记可观测日志并 Ack 消费路径，与现有 skip 语义一致）。
- 各厂商通知**标题行**仍为「胖宝」，本变更不改 logo / 大图 / pending API 结构。

## Capabilities

### New Capabilities

- `predict-imminent-push-copy`：预测临近可见推送正文模板、昵称回落、空 title 跳过发送。

### Modified Capabilities

- （无）`predict-imminent-offline-push` 相关规格尚未归档入 `openspec/specs/`；本变更以新 capability 固化文案与跳过语义，不改调度/去重/历史闸契约。

## Impact

- **代码**：`internal/services/voice/predict_imminent.go`（`handlePredictImminentFire` 拼装 alert）；经已有 `DeviceProfile()` 读 `babyName`。
- **API**：不改 `PUT /device/api/predict/imminent/pending` 请求结构；客户端须继续传非空 `title` 才能收到推送。
- **push-service**：标题「胖宝」与厂商 payload 不变；仅消费方传入的 `alert` 字符串变化。
- **客户端**：无强制发版；空 title 的旧同步将不再出通知（符合产品选择）。
