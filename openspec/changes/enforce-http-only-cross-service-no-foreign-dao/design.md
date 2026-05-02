## Context

仓库已具备 `device/profile_adapter.go` 的 **local | remote | canary** 画像访问模式，以及 `contracts.ResolveHTTPTargets()`。`voice` 意图与 DeepSeek 链路仍通过 `DeviceAdmin()` 进程内调用访问 `user`/`event`/`action` 表，在 **voice 进程 default DB = voice 库** 时会语义错误。目标部署为严格单服务单库，跨域只走 HTTP。

## Goals / Non-Goals

**Goals:**

- voice 对 device 域的读写在代码结构上 **仅经 HTTP 客户端**（失败语义、日志、可选 failover 与画像适配一致）。
- 保留迁移期开关：显式 `local` 仅用于开发或临时同库，默认生产为 `remote`。
- 与 `align-service-db-boundaries-history-voice-device` 已做的 history 委派相兼容。

**Non-Goals：**

- 不新增测试文件（仓库阶段约束）。
- 不改变 device 表结构；不在本变更中实现完整 gRPC，仅 HTTP。

## Decisions

1. **统一适配器入口**  
   - **决策**：扩展与 `DeviceProfileContract` 平行的 **DeviceAdmin 访问抽象**（或按能力拆成 `DeviceEventsClient`、`DeviceActionsClient`），底层实现 `localHTTP` / `remoteHTTP`。  
   - **备选**：在 `voice` 内散落 `http.Get`——拒绝，难以统一错误码与重试。

2. **local 含义**  
   - **决策**：`local` 仅表示「调用本机 URL」（如 `http://127.0.0.1:9803`），**不是**进程内 DAO；单库生产禁用「进程内 DAO 捷径」。  
   - **备选**：保留进程内 DAO 作 local——与单库目标冲突，仅作文档禁止。

3. **API 覆盖面**  
   - **决策**：列出 voice 当前调用的 `DeviceAdmin` 方法，逐一映射到已有或新增 **device internal HTTP**（与 admin 口令隔离的内部路由）。  
   - **备选**：全部走 admin 带口令 API——不适合语音链路自动化调用。

4. **配置**  
   - **决策**：复用 `DEVICE_SERVICE_URL`；引入或复用 `DEVICE_ADMIN_ACCESS_MODE`（命名可与 profile 对齐）控制 local/remote/canary。  
   - **备选**：每类资源单独 URL——配置爆炸，暂缓。

## Risks / Trade-offs

- **[Risk]** 调用链变长、延迟上升 → **缓解**：连接池、合理超时、缓存热点（事件/动作列表）。  
- **[Risk]** 双写与最终一致 → **缓解**：保持 device 为权威；voice 不缓存他域持久化状态。  
- **[Risk]** 遗漏的隐式 DAO import → **缓解**：评审清单 + 可选静态 grep 规则。

## Migration Plan

1. 实现 HTTP 客户端与 device 侧路由补全。  
2. voice 配置改为 `remote`（或 local=本机 device URL）。  
3. 验证语音全流程与管理端。  
4. 回滚：环境变量切回旧模式（若仍保留兼容分支）。

## Open Questions

- device 内部 API 是否统一前缀 `/device/internal/api/...` 并与网络策略（仅集群内）绑定？  
- 是否与 history 一样对 `ListEvents` 类只读接口增加短时 Redis 缓存（voice 侧）？
