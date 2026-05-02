## Context

当前代码库在拆分前按「单库多表」组织，`history` 包内 `localService` 同时操作 `dao.History`、`dao.User`、`dao.Event`、`dao.Suggest`；`voice` 侧仍存在对 `dao.Event`、`dao.Action` 的写入。生产目标库划分为：history-server（`history`、`domain_outbox`）、device-server（`user`、`event`、`action`）、voice-server（`qa`、`suggest`）。约束来源：`AGENTS.md` 禁止跨服务直查他域表；异步投影与 outbox 已部分落地。

## Goals / Non-Goals

**Goals:**

- 每个进程的数据访问与其部署库一致；跨域数据仅通过 HTTP 契约（及既有事件/outbox 路径）。
- 明确 `domain_outbox` 的写入方与数据库连接组一致，避免 device 进程误连非 history 库写 outbox。
- 保持迁移期 `local|remote|canary` 与 failover 语义，关键路径打日志。

**Non-Goals:**

- 不新增测试文件（仓库阶段约束）。
- 不一次性重写全部 OpenSpec 归档规格 Purpose 字段。
- 不改变业务侧「生日/事件/建议」的对外产品语义，仅调整数据获取路径与归属。

## Decisions

1. **history-service 数据面收缩**  
   - **决策**：`history/local.go` 仅保留 `history` 表与 `domain_outbox`（及与二者同事务所需的代码路径）。  
   - **备选**：在 history 内保留 HTTP 客户端转发到 device/voice——可作为过渡，但持久化仍不得在 history 进程执行 `user/event/suggest` SQL。优先：**调用方（gateway 或 voice）直调 device/voice**，减少 history 作为「二传手」 unless 已有统一 BFF。

2. **生日与事件选项的权威域**  
   - **决策**：`user`（生日/性别）、`event`（事件字典）由 **device-service** 暴露 API；history 不再提供这些表的 DAO 实现。  
   - **备选**：经 gateway 聚合——可与现有路由并存，但契约上仍以 device 为权威。

3. **建议（suggest）**  
   - **决策**：`suggest` 表仅由 **voice-service** 访问；history 的 `ListSuggest`/`DeleteSuggest` 从 local DAO 移除，改为 voice 契约或删除 history 侧重复 API（由路由切到 voice）。

4. **voice 对 event/action**  
   - **决策**：`voice_chat_understanding` 等对 `dao.Event`、`dao.Action` 的 Insert/查询改为 **device-service HTTP 客户端**（或已存在的 device 适配层），与 `voice-device-profile-http-contract` 一致。  
   - **备选**：异步消息——延迟更高，本次以同步契约为主。

5. **outbox 写入**  
   - **决策**：若 `domain_outbox` 仅存在于 history 库，则 device 侧 `enqueueDeviceProjectionEvent` 不得使用 `dao.User.Group()` 写 outbox；改为向 **history 服务**投递事件、或使用 **消息队列**、或经 **共享 outbox 库连接（仅当运维显式配置同库）**。默认推荐：**事件发布到 MQ / 或 history 提供的内部 outbox 接入 API**，与设计阶段与运维对齐。

6. **history 远程适配器**  
   - **决策**：`historyRemoteClient` 中指向 `/device/history/api/suggest`、`.../birthday`、`.../event/options` 的路径需迁移为 **voice / device 基址**；`Contract` 可拆分为多个接口或保留 façade 但实现改为多后端（避免误导「全是 history」）。

## Risks / Trade-offs

- **[Risk]** 拆分调用链增加延迟与故障点 → **缓解**：超时、熔断、迁移期 failover、缓存（现有 Redis 读模型）保留。  
- **[Risk]** 网关路由与客户端 URL 配置遗漏导致 404 → **缓解**：contracts 集中路径、文档与启动检查。  
- **[Risk]** 生日保存与 outbox 原同一事务 → **缓解**：改为最终一致（设备写成功 + 异步领域事件）；明确幂等与重试。  
- **[Risk]** 单体开发环境仍用单库多表 → **缓解**：环境变量区分「单库」与「分库」；local 模式文档说明。

## Migration Plan

1. 实现 device/voice 契约客户端与 controller 补全（若缺）。  
2. 切换 gateway 与内部调用顺序：先读新路径，再下线 history 上的越界路由。  
3. 数据无需迁移表间（表已在正确库）；仅需停写错误进程上的错误表。  
4. 回滚：恢复旧路由与适配器开关（feature flag / 环境变量）。

## Open Questions

- `domain_outbox` 在目标部署中是否**仅**挂在 history 库？若存在多副本，需确认中继进程连接组。  
- 对外 API 是否仍统一走 `/device/history/...` 前缀（网关重写），还是拆成 `/device/...` 与 `/voice/...` 公开路径。
