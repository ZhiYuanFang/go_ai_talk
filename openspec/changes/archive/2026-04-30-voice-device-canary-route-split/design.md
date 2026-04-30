## Context

当前 gateway 在 voice/device 路由上采用同一文件中的双分支实现，且仅支持 `local|proxy`。相比之下，history 已具备独立路由组件与 `canary` 分流能力，发布策略更灵活。为统一运维心智与发布能力，需要将 voice/device 的路由治理与 history 对齐。

本次变更关注 gateway 路由层，不涉及 voice/device 业务逻辑本身。核心目标是实现“按领域独立管理 + canary 分流 + 结构解耦”。

## Goals / Non-Goals

**Goals:**
- voice 与 device 分别使用独立路由代理中间件，不再共用一个中间件文件管理两域逻辑。
- voice 与 device 路由模式从 `local|proxy` 升级为 `local|proxy|canary`。
- 引入稳定分流键与百分比配置，保证 canary 流量一致性。
- 同步 compose 与 kustomize 配置，使各环境可显式控制 voice/device canary。

**Non-Goals:**
- 不改动 voice-service 与 device-service 的业务处理实现。
- 不在本次变更中引入动态配置中心或实时热更新路由策略。
- 不涉及 history 路由逻辑重写（仅对齐模式，不改变其现有机制）。

## Decisions

1. **按领域拆分路由中间件**
   - 决策：将 voice 与 device 路由代理拆分为独立中间件与配置结构。
   - 理由：降低文件耦合，避免一个改动影响两个领域。
   - 备选：保留单文件并继续扩展。
   - 不采纳原因：后续引入 canary 后逻辑复杂度进一步升高，维护风险大。

2. **voice/device 对齐 history 的三态路由**
   - 决策：新增 `canary` 模式，并保留 `local/proxy` 语义。
   - 理由：统一发布模型，便于按域灰度与回滚。
   - 备选：只保留 `local/proxy`。
   - 不采纳原因：无法满足分阶段放量需求，发布风险更高。

3. **采用稳定键哈希分流**
   - 决策：使用设备号（或等效稳定标识）+ 哈希的无状态分流方案。
   - 理由：同一设备请求命中稳定，减少灰度抖动。
   - 备选：随机百分比分流。
   - 不采纳原因：同一会话可能在 local/proxy 间跳转，影响一致性。

4. **配置显式化并跨环境同步**
   - 决策：为 voice/device 分别引入 canary 百分比与目标地址变量，并同步到 compose/kustomize。
   - 理由：保障开发、测试、部署环境行为一致。
   - 备选：仅在本地 compose 配置 canary。
   - 不采纳原因：环境差异会导致联调与发布行为不一致。

## Risks / Trade-offs

- **[路由配置项增多]** 变量数量提升，误配风险增加  
  → Mitigation：文档化默认值与合法取值，补充启动日志输出关键配置。

- **[canary 分流键不稳定]** 键选择不当会导致流量抖动  
  → Mitigation：优先使用设备号，缺失时回退到可重复计算键。

- **[拆分期间回归风险]** 路由路径可能出现漏配  
  → Mitigation：逐路径冒烟验证（voice text、device admin），并保留 `local` 回滚开关。

## Migration Plan

1. 拆分 voice/device 路由代理实现为独立文件与安装函数。
2. 为 voice/device 增加 canary 配置与分流函数。
3. 更新 compose 与 kustomize 的 gateway 环境变量。
4. 执行 local -> canary -> proxy 的分阶段验证。
5. 更新端点归属与运行契约文档，固化新规则。

回滚策略：
- 将 voice/device 路由模式回切到 `local` 或 `proxy`，并将 canary 百分比归零。
- 若拆分实现出现问题，可回滚到上一版本镜像恢复原路由逻辑。

## Open Questions

- voice/device canary 分流键是否统一使用 `deviceNo`，还是按路径定义独立键策略？
- 是否需要为 voice/device canary 命中率增加单独观测指标？
