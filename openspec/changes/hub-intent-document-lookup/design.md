## Context

九个飞轮的列表与删除已经在 `/voice/admin/api/*-vectors`。缺的是按原文一次查全。Python 变更 `cue-document-lookup` 提供聚合读接口。

## Goals / Non-Goals

**Goals:**

- Hub 输入 document，列出全部精确命中，并可删除其中一条。
- 删除仍走现有按维度的 DELETE。

**Non-Goals:**

- 在 Go 里分页扫描九个列表再过滤。
- 改各维度灌种页。
- 计入 App usage（本接口是 voice admin）。

## Decisions

### D1：只代理聚合读

- **选择**：`GET /voice/admin/api/intent-cues/by-document?document=`，voice 进程转发 Python，原样带回 `dimension`、`collection`、`id`、`document`、`payload`。
- **理由**：过滤发生在 Python 一次 get，避免 Go 对每个集合反复拉全量分页。

### D2：删除映射

- **选择**：门户按 `dimension` 调用已有删除：
  - `op` → `/voice/admin/api/op-vectors/{id}`
  - `intent` → `intent-vectors`
  - `qty` → `qty-vectors`
  - `time` → `time-vectors`
  - `remark` → `remark-vectors`
  - `read_source` → `read-source-vectors`
  - `predict` → `predict-vectors`
  - `history_speak` → `history-speak-vectors`
  - `locate` → `locate-vectors`
- **理由**：id 只在所属集合内唯一。

### D3：门户布局

- **选择**：查询条和命中表放在 `intent-agent-admin.html` 飞轮卡片之上或之下，表占满内容宽，一行一条命中。未知 `dimension` 不提供删除按钮。
- **理由**：运维从这一页出发，不必再进九个子页。

## Risks / Trade-offs

- [Python 未部署时查询失败] → 门户显示错误，不把失败显示成零命中以外的成功空表时要带上错误文案。接口非 2xx 即错误。

## Migration Plan

先上 Python，再上 voice 与静态页。未升级的 Python 会使该查询 404/500，各维度原页面仍可用。

## Open Questions

- （无）
