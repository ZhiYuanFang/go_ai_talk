## 1. leadSeconds 可读（history 侧）

- [x] 1.1 在 `internal/platform` 或 history 包内提供与 voice 对齐的 `leadSeconds` 读取（env `VOICE_PREDICT_IMMINENT_LEAD_SECONDS`，非法/空默认 300）；禁止 history import `services/voice`
- [x] 1.2 中文注释标明与普通推送 lead / MQ fireAt 同语义

## 2. Refresh 组装：进窗过滤 + 总条数 3 + 进行中优先

- [x] 2.1 在 `buildLiveActivityItems`：pending 过滤为 `nextAt - leadSeconds ≤ now`；保留 `nextAt` 升序
- [x] 2.2 实现截断：先取至多 3 条 open timing（`startTime` 倒序），剩余名额填进窗预测；发射顺序仍「预测段 + 进行中段」
- [x] 2.3 更新 `RefreshBabyLiveActivity` / `buildLiveActivityItems` 文件与函数中文注释（去掉「pending 全量」表述）
- [x] 2.4 确认未改 CUD Del→Refresh、pending PUT Refresh、MQ fire Refresh 触发点；未新增 ticker

## 3. Token 表去掉 attributes_kind / activity_id

- [x] 3.1 调整 `push_schema` ensure：新表不建这两列；已有表 DROP（并确认唯一键仍为 `(wx_id, device_key, token_kind)`）
- [x] 3.2 清理 `entityLiveActivityToken` / `RegisterLiveActivityToken` 等对两列的读写与恒空写入
- [x] 3.3 全文检索确认 push 域无残留业务依赖这两列分槽

## 4. 校验

- [x] 4.1 `gofmt` 相关改动；按需编译 history-service / push-service
- [x] 4.2 自检：远点 pending 不进 items；进行中≥3 时无预测项；2 timing + 多预测 → 3 条且含 1 条最近预测
