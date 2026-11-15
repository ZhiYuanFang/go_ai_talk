## ADDED Requirements

### Requirement: voice AI 额度与 lane_entitlement MUST 删除

voice-service MUST 删除 `ai_quota_*` 业务读写、App/Admin AI 额度 API、`lane_entitlement`（VIP∨额度∨硬件 → premium）及 care/growth 日限对 quota 的依赖。相关 Redis usage 键写入路径 MUST 停止。UCG 库中的 polish 额度删除见 `retire-ucg-polish`；本 requirement 覆盖 voice 库额度。

#### Scenario: voice App 额度 API 下线

- **WHEN** 客户端请求原 `GET /voice/app/api/ai-quota`
- **THEN** 服务端 MUST NOT 再返回 voiceAi/clinicAi/careAlert 额度业务响应（路由已删除或未注册）

### Requirement: llm_lane_config 与 Hub 选模 MUST 删除

系统 MUST 删除 voice `llm_lane_config` 的 Ensure/Admin GET-PUT、Hub「AI 模型与并发」中 Voice/UCG/Sim lane 编辑能力，以及业务请求向 Python/上游携带 provider/model 的选模逻辑。运维 MUST NOT 再通过该页配置理解/clinic/care/polish/sim 模型。

#### Scenario: llm-lanes Admin API 下线

- **WHEN** 运维请求原 `/voice/admin/api/llm-lanes` 或等价选模 Admin API
- **THEN** 服务端 MUST NOT 再提供可写的多 lane 选模配置接口

### Requirement: aimodel 多 lane 中台 MUST 退役

在 clinic/care/polish/sim/understanding 选模调用方删除后，系统 MUST 删除或使 `internal/services/aimodel` 不再被业务进程引用（含 `Acquire`/`Invoke*`/`ProfileStore` 注册）。`media_gen` 等仅 sim 使用的生成封装 MUST 随 sim 删除。

#### Scenario: 业务包不再 import aimodel

- **WHEN** 对 `internal/services/voice`、`internal/services/ucg`、`internal/services/simuser`（若仍存在）做 import 检查
- **THEN** MUST NOT 再存在对 `hello/internal/services/aimodel` 的业务引用（包本身可删除）
