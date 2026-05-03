# 仓库级 AI 执行约定

## 文档语言
- OpenSpec 相关文档（proposal/design/specs/tasks）默认使用中文说明性文本。
- 环境变量、路径、接口、协议关键字、代码符号可保留英文原文。

## 代码注释
- 新增或修改代码时，需尽可能补充中文注释，优先覆盖复杂流程、关键边界条件、跨服务调用与错误语义。
- 公共导出函数/结构体/关键字段优先补充中文注释。
- 禁止无意义注释，注释应解释“为何这样实现”和“失败时行为”。

## 服务边界（强制）
- 跨服务数据访问必须走服务接口契约（HTTP/RPC/事件），禁止在服务实现中直接访问他域 DAO 或数据库表。
- 迁移期若使用 `local|remote|canary` 双路径，必须保留显式配置与 failover 语义，并输出可观测日志。
- 代码评审时需显式检查：是否出现跨库直查、是否补齐契约路径与错误语义。
- `internal/services/history` 不得 import 或调用 `dao.User`、`dao.Event`、`dao.Suggest` 等他域表；`device-service` 写 `domain_outbox` 须使用 `database.history_relay`（或经评审的等价投递），禁止用 `dao.User` 默认连接组误连。
- `internal/services/voice` 不得 import `hello/internal/dao` 中 **user/event/action** 等他域表 DAO；设备域访问 MUST 经 `voice.DeviceAdmin()`（HTTP 实现）或经批准的契约，禁止在 voice 进程内直连 device 库表。评审可检索：`grep -r \"internal/dao\" internal/services/voice`（应仅出现 suggest/qa 等本域表）。
- 服务默认配置必须按进程独立（`gateway`/`voice-service`/`device-service`/`history-service`），禁止回退到共享主配置承载他域业务项。
- `manifest/config/config.yaml` 仅允许保留网关与全局公共配置；评审时必须检查是否有 voice/device/history 专属字段回流。
- `voiceChat`（ASR/LLM/TTS/会话）必须维护在 `manifest/config/voice-chat.shared.yaml`，供 `voice-service` 与 `history-service` 共用；禁止在 `config.voice-service.yaml` 与 `config.history-service.yaml` 中重复整段 `voiceChat`（迁移期可仅依赖 `GF_GCFG` 中的 `voiceChat` 作兜底，不作为长期双源）。
- 代码目录边界与包边界必须一致：业务实现统一位于 `internal/services/**`，禁止新增实现文件到 `internal/service`。
- 评审时必须检查 import 路径，确保 `cmd`/`controller` 不再依赖 `hello/internal/service` 旧路径。

## 测试文件
- 当前阶段不新增测试文件（包括 `*_test.go`、`*.spec.*`、`*.test.*`）。
