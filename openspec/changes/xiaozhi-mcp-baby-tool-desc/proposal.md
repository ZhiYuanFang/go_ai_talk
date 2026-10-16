## Why

小智 MCP 工具 `baby_feeding_advisor` 的描述以喂养记录为主，宝宝名字、月龄、睡眠、发育等话题容易被小智自己回答，进不了现有意图智能体。只需放宽工具描述的选路范围，不改调用链。

## What Changes

- 重写 `internal/services/mcpbridge/tools.go` 中 `chatToolDescription`：定位为宝宝顾问；用户谈到宝宝时调用本工具，不限于喂养；保留喂养记录/查询、宝宝信息（名字、生日、月龄、性别、身高体重、睡眠、发育与日常照护）、以及对上一条本工具操作的确认/取消短句。
- 文案用「调用」，不再堆叠多句「必须」。
- 工具名 `baby_feeding_advisor`、入参 schema、`ChatHandler`、`/voice/chat/ws` 路径保持不变。
- **不**新增 MCP 工具，**不**改 Python 意图实现，**不**改额度或开通闸门。

## Capabilities

### New Capabilities

- `xiaozhi-mcp-baby-tool-desc`：放宽 `baby_feeding_advisor` 的 tools/list 描述，使宝宝相关话题（不限于喂养）仍由该工具进入现有智能体。

### Modified Capabilities

- （无。基线见 `openspec/specs/v3.0.3/spec.md` 中 mcp 工具经 `/voice/chat/ws` 文模式的既有 Requirement；本变更只改描述文本，不改该执行契约。）

## Impact

- **代码**：`internal/services/mcpbridge/tools.go` 的 `chatToolDescription`。
- **运行**：已连接音箱需重连后才会拿到新的 `tools/list` 描述；无 API、无 DB、无发版清单变更。
- **usage**：不新增 App HTTP 路由。
