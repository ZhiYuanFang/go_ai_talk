## 1. API 与载荷

- [x] 1.1 扩展 `PredictImminentPendingEventItem`：可选 `expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit` / `expectLeafTitle`（同路径 omitempty，不新建 v2）
- [x] 1.2 扩展 `PredictImminentPendingItem` 与 `predictImminentFirePayload`，controller 映射透传
- [x] 1.3 `SyncPredictImminentPending` 整表写入 Redis 时保留 expect；PublishDelayed 载荷复制 expect
- [x] 1.4 `sync_recv` 摘要附带 expect（kind + durationSec 或 amount），便于联调 grep

## 2. 推送文案

- [x] 2.1 在 `handlePredictImminentFire` 保留基句 `{昵称}要{title}了`；expect **不**写入 push `data`
- [x] 2.2 实现 `formatExpectDurationCN(sec)`（design D3：小时浮点 + round；覆盖 59/90/3600/3900 等边界）
- [x] 2.3 合法 duration expect 追加 `，预计这次约` + 中文时长
- [x] 2.4 合法 amount expect 追加 `，预计这次约` + 数量 +（可选）unit +（可选）leafTitle；空 unit 仍可拼数量
- [x] 2.5 按 D1 矩阵处理非法/未知 kind；`expectLeafTitle` trim 后截断 32 字

## 3. 验收

- [x] 3.1 用含 duration expect 的 pending PUT → Redis/MQ 可见字段 → 消费路径拼出带时长后缀的 alert（可日志验证）
- [x] 3.2 用含 amount+叶子 的 pending PUT → alert 形如「…要喂养了，预计这次约120毫升配方奶」
- [x] 3.3 旧客户端无 expect 字段 PUT → 行为与现网一致（仅基句）
- [x] 3.4 与 Flutter 同名 change 字段对齐联调一次
