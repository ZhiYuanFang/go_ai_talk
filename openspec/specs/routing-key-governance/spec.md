# routing-key-governance Specification

## Purpose
定义路由键治理约束，确保事件发布链路仅允许已注册路由键并集中管理路由定义。

## Requirements
### Requirement: 路由键白名单校验
系统 MUST 对 outbox 写入与事件发布链路中的 `routing_key` 执行白名单校验，未注册路由键不得进入发布流程。

#### Scenario: 合法路由键正常发布
- **WHEN** 业务写入或发布已注册的 `routing_key`
- **THEN** 系统 MUST 允许进入 outbox 与发布流程

#### Scenario: 非法路由键被拒绝
- **WHEN** 业务使用未注册的 `routing_key`
- **THEN** 系统 MUST 拒绝该请求并输出结构化错误日志

### Requirement: 路由键集中定义
系统 SHALL 提供集中路由键定义与查询接口，避免在多个模块重复硬编码路由字符串。

#### Scenario: 新增路由键需要注册
- **WHEN** 开发新增事件路由键
- **THEN** 该路由键 MUST 先在集中注册处声明后才能被调用方引用
