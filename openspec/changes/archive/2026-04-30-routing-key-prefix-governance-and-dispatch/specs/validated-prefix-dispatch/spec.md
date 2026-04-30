## ADDED Requirements

### Requirement: 分发前必须执行路由键白名单校验
系统 MUST 在 outbox 发布与投影分发入口先执行路由键合法性校验；未注册路由键必须被拒绝并记录拒绝来源。

#### Scenario: 收到未注册路由键
- **WHEN** outbox 处理到不在注册表中的 `routing_key`
- **THEN** 系统必须拒绝该事件并输出包含来源模块的告警日志

### Requirement: 校验通过后必须按前缀分组分发
系统 SHALL 在路由键通过合法性校验后，基于前缀常量将事件分发到对应领域处理器，而非依赖逐项路由键枚举分支。

#### Scenario: 处理 history.record 前缀事件
- **WHEN** 事件路由键为 `history.record.*` 且已通过白名单校验
- **THEN** 系统必须将该事件分发给 history 投影处理器

### Requirement: 必须提供未知前缀默认保护
系统 MUST 为已注册但未映射分发处理器的前缀保留默认保护分支，避免静默忽略。

#### Scenario: 路由键合法但前缀未绑定处理器
- **WHEN** 事件通过白名单校验但其前缀没有配置分发处理器
- **THEN** 系统必须输出告警并按既定失败语义处理
