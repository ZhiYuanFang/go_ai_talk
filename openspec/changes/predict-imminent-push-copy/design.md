## Context

预测临近离线提醒已在 `voice-service` 落地：客户端 `PUT` 全量同步 Redis 待办，延时 MQ 叫醒后校验去重与历史闸，再经 `clients/push` 以 `bizType=predict_imminent` 扇出。可见正文现由 `handlePredictImminentFire` 拼为 `宝宝提醒：` + `title`（空 title 时回落「事件将在约 N 分钟内发生」）。各厂商通知标题硬编码「胖宝」。device 画像契约 `DeviceProfile().GetProfile` 已在 voice 其他路径使用，可直接取 `babyName`。

约束：跨域只走 `clients/*`；不改 v1 pending API 结构；不新增测试文件；消费路径业务 skip 一律 Ack。

## Goals / Non-Goals

**Goals:**

- 可见推送正文统一为 `{昵称}要{title}了`。
- 昵称为空或画像失败时用「宝宝」；`title` 为空则不推。
- 保持调度、去重、历史闸、扇出与标题「胖宝」不变。

**Non-Goals:**

- 事件 logo / 大图；按事件树展开到叶子名。
- 修改 `PUT .../pending` 字段或新开 v2。
- 修改 push-service 默认标题或厂商 image 字段。
- 客户端强制发版或改预测算法。

## Decisions

### D1. 事件名只用客户端 `title`

- **选择**：模板中的事件名 = MQ/Redis 载荷中的 `title`（trim 后）。
- **理由**：客户端同步的是父事件展示名；服务端猜叶子成本高且易错。
- **备选**：`ListEvents` 按 `eventId` 查字典 `name` — 本变更不采用（与产品「用客户端 title」一致）。

### D2. 空 `title` 跳过推送

- **选择**：trim 后为空则记 warning 日志、不调用 `PushByBizType`，handler 仍返回 nil（Ack）。
- **理由**：避免「宝宝要事件将在约 5 分钟内发生了」类怪句；产品明确选择不推。
- **备选**：回落「发生一件事」— 未采用。
- **去重键**：若在拼装前已占用 5 分钟去重键，空 title 跳过仍占用去重窗口（与现有「进入发送前占位」一致）；实现时保持与当前 `SetNX` 时序一致，避免为文案失败额外清键（可接受短窗内不重试同节点）。

### D3. 昵称来源与回落

- **选择**：推送前 `DeviceProfile().GetProfile(deviceNo)`；`BabyName` 非空则用，否则「宝宝」；GetProfile 失败同样回落「宝宝」并打 warning。
- **理由**：与产品一致；失败不阻断提醒。
- **备选**：把昵称塞进 pending — 会扩 API 且改名后需重同步，不做。

### D4. 标题行不动

- **选择**：HMS/MiPush 等仍用「胖宝」；本变更只改传入的 `alert` 正文。
- **理由**：产品明确标题行保持品牌名。

## Risks / Trade-offs

- [客户端偶发不传 title] → 该节点不推；日志可观测；CLIENT 文档可提醒 title 必填语义（非 API 破坏性字段变更）。
- [画像接口抖动] → 回落「宝宝」，仍推；不放大失败面。
- [去重键在空 title 前已占] → 五分钟内同节点不会因补传 title 自动再推，须等下次 pending 刷新换 nextAt 或去重过期；可接受。

## Migration Plan

1. 部署 voice-service 含文案逻辑即可；无 DB/配置迁移。
2. 回滚：还原 `handlePredictImminentFire` 旧 alert 拼装。
3. 可选：更新 `predict-imminent-offline-push/CLIENT.md` 或本变更旁注，强调 title 非空才会推送。

## Open Questions

- （无）产品决策已闭合：模板、昵称回落、空 title 不推、标题「胖宝」不动。
