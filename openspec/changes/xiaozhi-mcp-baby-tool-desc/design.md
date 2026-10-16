## Context

`baby_feeding_advisor` 由 `xiaozhi-mcp-service` 在 `tools/list` 中暴露。小智按描述决定是否把用户原话放入 `transcript` 并 `tools/call`。`ChatHandler` 校验开通后经 `ChatViaVoiceWS` 打 `/voice/chat/ws` 文模式。现描述编号几乎全是喂养，总括句又堆了多处「必须」。

## Goals / Non-Goals

**Goals:**

- 描述覆盖宝宝相关信息，不限于喂养记录。
- 确认/取消短句仍指向同一工具，避免澄清断轮。
- 语气用「调用」，避免再堆「必须」。

**Non-Goals:**

- 不改工具名、input schema、Handler、WS 协议。
- 不新增工具，不分流到 `/voice/clinic/ws`。
- 不改 Python 意图能否回答档案问题（描述只影响小智是否把原话送进来）。

## Decisions

1. **只改 `chatToolDescription` 字符串**
   - 文案结构：标题「宝宝顾问」+ 一句范围（谈到宝宝时调用，不限于喂养）+ 三条：喂养、宝宝信息、确认/取消 + 一句「传入原话由工具理解并回答」。
   - **备选**：改工具名 — 小智侧已按现名登记，否决。
   - **备选**：加第二工具 — 与「单工具续聊」相反，否决。

2. **生效方式**
   - 新描述随进程发布进入 `tools/list`。已建立的小智会话可能缓存旧描述，需重连 Bridge 后生效。本变更不自动踢连接。

## Risks / Trade-offs

- [小智仍偶发不调用] → 描述给例子，不保证 100% 选路；重连后用真实语句抽测。
- [智能体答不出档案题] → 本变更不修 Python；若上线后仍答非所问，另开意图侧变更。

## Migration Plan

- 发布 `xiaozhi-mcp-service` 后，对已绑定音箱执行一次重连（App/Hub 已有 reconnect）以刷新工具列表。
- 回滚：还原描述字符串。

## Open Questions

- （无。）
