## Why

运维在意图智能体门户要输入一整句原文，看它出现在哪些飞轮维度，并直接删掉某一条。现在只能进单个维度翻页，对不上一句 document。

## What Changes

- voice-service 新增管理读接口，转发 Python `GET /v1/admin/cues/by-document`（兄弟仓 `cue-document-lookup`）。
- 意图智能体门户增加原文输入、查询，以及命中表：维度、向量 id、payload、删除。
- 删除调用该维度已有的 `DELETE /voice/admin/api/*-vectors/{id}`，不新增跨集合删除。
- 不查类型向量、规则缺口、覆盖缺口。查询为空时不发请求。
- 结果表近全宽，每人/每条命中一行，方便看 payload。

## Capabilities

### New Capabilities

- `hub-intent-document-lookup`：门户按 document 查看九个飞轮命中并按维度删除。

### Modified Capabilities

- （无）

## Impact

- **voice-service**：新的 admin GET，Admin 口令与现有向量管理一致。`/voice/admin/api/*` 已不计入 App usage。
- **Hub**：`resource/public/intent-agent-admin.html`。
- **Python**：依赖已部署的 `cue-document-lookup`。本仓不改 Chroma。
