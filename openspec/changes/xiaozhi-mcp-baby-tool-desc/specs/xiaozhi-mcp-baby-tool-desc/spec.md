## ADDED Requirements

### Requirement: 宝宝话题经同一 MCP 工具进入现有智能体

`xiaozhi-mcp-service` 暴露的工具 `baby_feeding_advisor` 的 description MUST 表明：用户谈到宝宝时调用本工具，范围不限于喂养记录。描述 MUST 覆盖喂养记录或查询、宝宝信息（至少包括名字、生日、月龄、性别、身高体重、睡眠、发育与日常照护），以及用户对上一条本工具操作的确认或取消短句（原话作为 `transcript`）。描述 MUST NOT 以多句「必须」堆叠选路。工具名、入参与执行路径 MUST 保持为既有 `ChatHandler` → `/voice/chat/ws` 文模式，MUST NOT 新增 MCP 工具。

#### Scenario: tools/list 描述含宝宝信息且不限喂养

- **WHEN** 客户端请求 `tools/list`
- **THEN** `baby_feeding_advisor` 的 description 包含宝宝信息示例，并说明调用范围不限于喂养
- **AND** 仍为唯一聊天工具名 `baby_feeding_advisor`

#### Scenario: 确认短句仍走同一工具

- **WHEN** 阅读该 description 中关于确认或取消的条目
- **THEN** 短句示例仍要求调用本工具并把用户原话放入 `transcript`
