## 1. 契约与路由

- [x] 1.1 梳理 `internal/services/contracts` 与 gateway 路由：为 device（用户画像、事件、动作）与 voice（suggest）补充或校正 HTTP 目标，避免误挂在 history baseURL 下。
- [x] 1.2 明确对外路径策略（保留 `/device/history/...` 聚合 vs 拆分），并更新 `docs/runbooks` 中环境变量说明。

## 2. history-service 收敛

- [x] 2.1 从 `internal/services/history/local.go` 移除对 `dao.User`、`dao.Event`、`dao.Suggest` 的读写；`SaveBirthday`/`GetDeviceBirthday`/`ListEventOptions`/`ListDeviceSuggest`/`DeleteDeviceSuggest` 改为调用 device/voice 客户端或删除并由调用链上移。
- [x] 2.2 调整 `history/adapter.go` 中 remote 客户端：非 history 数据改指向 device/voice URL；`switchAdapter` 与缓存失效语义与真实下游一致。
- [x] 2.3 更新 `cache_rebuild.go`：`RebuildHistoryMetaCache`、`RebuildBirthdayCacheByDevice` 等改为从正确服务拉取或废弃误用 DAO 的路径。
- [x] 2.4 校验 `insertOutboxEventTx` 仅与 `history`/`domain_outbox` 同事务；移除对 `dao.User.Group()` 与跨表事务的假设。

## 3. device-service

- [x] 3.1 审查 `internal/services/device/admin.go` 中 `enqueueDeviceProjectionEvent`：若 `domain_outbox` 仅在 history 库，改为 MQ、history 内部 API 或经评审的同库配置，禁止错误连接组写 outbox。
- [x] 3.2 确认 device 暴露的 HTTP 已覆盖生日保存、事件选项、动作/事件写入等被 voice/history 迁移过来的调用点。

## 4. voice-service

- [x] 4.1 将 `voice_chat_understanding.go`（及关联文件）中对 `dao.Event`、`dao.Action` 的访问替换为 device HTTP 客户端/适配层。
- [x] 4.2 确认 `qa`、`suggest` 仅访问 voice 库；移除任何对 history 本地 suggest 路径的依赖。

## 5. 验证与文档

- [x] 5.1 在分库配置下手工验证：history 仅连 history 库、device 连 device 库、voice 连 voice 库的关键路径（启动与核心 API）。
- [x] 5.2 更新架构/评审检查清单：禁止 history 包 import 他域 DAO；device 不写错组 outbox。
