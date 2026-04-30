## Context

当前网关在 `/voice/chat/ws` 路径上仍执行本地语音 WS 业务，而 `voice-service` 已具备同等 WS 处理能力，导致入口职责重叠。项目目标态要求 gateway 仅承担流量与策略层能力，因此需要将 WS 路径改为边缘透传，并移除 gateway 本地 WS 业务执行。

该改动涉及协议升级链路（HTTP -> WebSocket）、容器网络内目标地址解析、运行时路由开关与失败语义，需要在不修改前端连接地址的前提下完成迁移。

## Goals / Non-Goals

**Goals:**
- gateway 支持 `/voice/chat/ws` 可配置透传到 `voice-service`。
- gateway 移除本地 `registerVoiceChatWS` 业务绑定，避免双执行。
- 保持前端对外 WS 入口地址不变（仍连 gateway）。
- 明确代理失败语义：目标不可达时返回握手/代理错误，不隐式回落本地。
- 通过环境变量支持迁移窗口内的模式切换（`local|proxy`）。

**Non-Goals:**
- 不改动 `voice-service` 内部 WS 业务逻辑。
- 不引入新的消息协议字段或鉴权机制改造。
- 不在本变更中实现跨集群的 WS 全链路负载均衡策略。

## Decisions

1. **采用显式 WS 路由模式开关**
   - 决策：新增 `VOICE_WS_ROUTE_MODE`（`local|proxy`）与 `VOICE_WS_PROXY_URL`。
   - 理由：便于灰度切换与快速回滚。
   - 备选：固定全量 proxy。
   - 不采纳原因：迁移窗口内缺少回退抓手。

2. **gateway 删除本地 WS 业务绑定**
   - 决策：在 gateway 路由注册中移除本地 `registerVoiceChatWS`，改由透传中间件承接。
   - 理由：保证 gateway 只做边缘层，不承载领域执行。
   - 备选：保留本地绑定并在代理失败时回退本地。
   - 不采纳原因：会形成双语义与不可预测行为，违背收敛目标。

3. **失败语义保持显式失败，不自动降级**
   - 决策：当处于 proxy 模式且目标不可达时直接返回握手/代理错误。
   - 理由：与项目 fail-fast 原则一致，避免掩盖下游故障。
   - 备选：自动回退本地 WS。
   - 不采纳原因：会隐藏配置/网络问题，增加排障成本。

4. **保持前端外部入口不变**
   - 决策：前端继续使用 gateway 的 `/voice/chat/ws`。
   - 理由：减少前端与网关同时变更带来的联调风险。
   - 备选：前端直接改连 `voice-service`。
   - 不采纳原因：需要客户端发布配合，迁移成本更高。

## Risks / Trade-offs

- **[WS 代理实现复杂度上升]** 连接升级与错误处理边界更复杂  
  → Mitigation：限定首版只支持单路径透传，补充冒烟校验脚本与日志字段。

- **[目标服务不可达导致前端可见失败]** proxy 模式下会直接暴露错误  
  → Mitigation：在编排中增加 health 依赖与上线前连通性校验。

- **[迁移窗口配置漂移]** 不同环境可能出现 mode/url 不一致  
  → Mitigation：将变量写入 compose 与部署清单，并在运行文档中提供标准值。

## Migration Plan

1. 在 gateway 新增 `/voice/chat/ws` 透传组件与配置解析。
2. 在 gateway 路由注册中移除本地 `registerVoiceChatWS`。
3. 在容器编排与部署配置中加入 `VOICE_WS_ROUTE_MODE`、`VOICE_WS_PROXY_URL`。
4. 在开发/测试环境以 proxy 模式联调，验证握手、消息收发、断连行为。
5. 上线时先灰度环境验证，再全量切换为 proxy。

回滚策略：
- 将 `VOICE_WS_ROUTE_MODE` 回切为 `local`，并恢复 gateway 本地 WS 绑定版本镜像。

## Open Questions

- WS 透传是否需要附加 header 白名单透传策略（例如 trace、auth）？
- 是否要为 WS 代理失败增加独立指标（握手失败率、断连率）并纳入告警？
