## Why

Hub 各维度飞轮页只能手写 JSON 种子，字段枚举散落在兄弟仓 Python 里，复制给其它 LLM 批量生成时容易漏字段或用旧结构。预测页示例仍缺 `predict_aspect`。需要在现有管理页放可复制提示词，并与 Python 载荷对齐。

## What Changes

- 门户已链的 9 个向量页（OP、事件、数量、时间、备注、R·数据源、R·predict、查史播报、U/D 定位）在批量上传区展示只读种子提示词，并提供一键复制。
- 提示词写明输出为 JSON 数组 `[{document, payload}]`，以及该维 Python 权威字段与枚举；事件维 document 为短说法，其余为整段原话。
- 预测页批量示例与提示词补上 `predict_aspect`（`when|amount|duration`）；门户卡片文案同步三字段。
- `type-vector-admin.html` 若保留提示词，须标明热路径已停、勿当现行灌种入口；**不**重新挂回意图门户。
- **不**新增 API、不改 Python、不新增 Chroma 集合。

## Capabilities

### New Capabilities

- `hub-intent-seed-prompts`：各维 Hub 向量页提供可复制的种子生成提示词，预测维示例对齐 Python 三字段。

### Modified Capabilities

- （无。基线见 `openspec/specs/v3.0.3/spec.md` 中既有飞轮管理页；本变更只加静态提示文案与预测示例字段。）

## Impact

- **静态页**：`resource/public/*-vector-admin.html`、`intent-agent-admin.html`。
- **部署**：gateway 静态资源；无后端发版。
- **usage**：不新增 App HTTP 路由。
