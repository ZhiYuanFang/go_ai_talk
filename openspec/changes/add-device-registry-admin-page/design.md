## Context
当前项目已具备文字与语音智能对话接口，并在内存中按 `deviceNo` 维护会话上下文；`user` 表已存在 `device_no`、`active_time`、`last_talk_time` 字段，但未形成完整的设备注册与可视化管理闭环。

本变更需要同时覆盖：
1) 新增网页交互（口令进入、注册设备、查看列表）；
2) 新增设备持久化写入；
3) 改造对话链路在成功后回写设备最近对话时间与最后一轮问答内容。

## Goals / Non-Goals
- Goals:
  - 提供最小可用的设备管理网页，满足“口令进入 + 注册 + 列表查看”。
  - 保证 `deviceNo` 全局唯一注册。
  - 保证每次智能对话成功后更新 `last_talk_time`、`last_talk_ask`、`last_talk_answer`。
- Non-Goals:
  - 不引入口令账户体系、JWT 登录页或复杂权限模型。
  - 不调整现有 DeepSeek/STT/TTS 业务策略。
  - 不在本次提案中引入设备删除、分页筛选、导出等增强能力。

## Decisions
- Decision 1: 页面口令采用固定值校验（`a521521521`）
  - Why: 用户明确要求固定口令，先满足最小需求。
  - Alternatives:
    - 配置化口令：灵活但超出当前要求。

- Decision 2: 设备注册与设备列表走后端 API，页面仅作展示与提交
  - Why: 便于后续扩展与测试，且符合现有控制器/服务分层。
  - Alternatives:
    - 纯静态文件直连数据库：不符合当前后端架构。

- Decision 3: 智能对话成功后再更新 `last_talk_time`
  - Why: 避免上游失败造成“伪活跃”时间。
  - Alternatives:
    - 请求到达即更新时间：实现简单但与“最后一次对话时间”语义不一致。

- Decision 3.1: 最后一轮提问与回答写入 `last_talk_ask` / `last_talk_answer`
  - Why: 满足“记录最后一次对话问答”需求，便于排障与运营查看。
  - Alternatives:
    - 仅存 `last_talk_time`：信息不足，无法定位最后一轮内容。

- Decision 4: 智能对话需设备先注册
  - Why: 用户新增设备管理能力的核心是“设备可控”；未注册设备直接对话会破坏管理边界。
  - Alternatives:
    - 对话时自动创建设备：弱化注册流程价值，且与“在网页上注册设备号”目标冲突。

## Risks / Trade-offs
- 固定口令存在泄露风险，且无法按环境区分。
  - Mitigation: 在后续迭代中改为配置项并支持环境隔离。
- 对话链路新增数据库写操作，可能增加延迟。
  - Mitigation: 仅在成功路径执行单条更新，失败不重试；必要时后续异步化。
- `last_talk_ask` / `last_talk_answer` 可能较长。
  - Mitigation: 字段长度按数据库约束设计，必要时截断并在规范中明确策略。
- 未注册设备被拒绝可能影响既有调用方。
  - Mitigation: 在 README/接口文档中明确前置注册要求，并提供可视化注册入口。

## Migration Plan
1. 增加设备管理页面与 API（不影响既有接口）。
2. 上线设备注册后，再启用“未注册设备禁止对话”校验（可灰度）。
3. 对话链路接入 `last_talk_time`、`last_talk_ask`、`last_talk_answer` 更新。
4. 通过回归测试确认文字/语音主流程兼容。

## Confirmed Scope
- `/voice/chat/ws` 的“每一次智能对话”按一次 `start...end` 成功结果为更新时间与最后问答内容写入粒度。
- 未注册 `deviceNo` 调用智能对话时直接拒绝。

## Open Questions
- `active_time` 与 `last_talk_time` 的存储格式是否继续沿用当前 UTF8 字符串时间，还是迁移为标准时间类型（本次维持现状）。
