## ADDED Requirements

### Requirement: careAlert 车道参与并发闸门

`careAlert` lane MUST 注册到与其它 lane 相同的 aimodel 并发闸门机制；`maxInFlight`/`maxWaiters` MUST 生效。同一 `provider+model` 跨 lane 共享池的既有语义 MUST 保持（含 free 模型若指向相同 provider+model）。

#### Scenario: careAlert 队列满

- **WHEN** careAlert 闸门等待队列已满且新的 premium 生成请求需要 Acquire
- **THEN** 服务 MUST 按既有 50301（或该路径等价错误）拒绝，MUST NOT 调用 Python 分析，MUST NOT consume `care_alert`
