## ADDED Requirements

### Requirement: 胖宝诊疗 clinic MUST 下线

系统 MUST 删除 `/voice/clinic/ws`（及等价 clinic HTTP）、`clinic_*` 服务实现、`clinic_ai` 额度字段与 Admin/App 展示、LaneClinic 配置，以及 Go 对 Python `/v1/clinic/stream` 的调用。gateway 对 clinic WS 的反代与鉴权豁免（若仅服务 clinic）MUST 同步清理。

#### Scenario: clinic WS 不可用

- **WHEN** 客户端尝试连接原 clinic WebSocket 路径
- **THEN** 服务端 MUST NOT 再注册 clinic 业务处理器

### Requirement: 聊天成长建议 MUST 删除

系统 MUST 删除喂养聊天中的「成长建议」生成路径（含 `callDeepSeekGrowthSuggestion` 及仅为其服务的 payload/意图分支），MUST NOT 再调用 Python `POST /v1/clinic` 同步陪伴接口。Hub/history 若仅用于运维读删历史 `suggest` 行，MUST 删除或停止暴露「成长建议」生成相关入口；历史行 MAY 留库。

#### Scenario: 不再触发成长建议 LLM

- **WHEN** 用户走喂养语音聊天主路径
- **THEN** 系统 MUST NOT 调用成长建议专用生成函数，MUST NOT 请求 `/v1/clinic`
