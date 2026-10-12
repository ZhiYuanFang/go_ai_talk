## Context

兄弟仓 `intent-crud-other-reframe` 已将意图热路径切到 OP 轨（create|read|update|delete|exit|other）+ 分层整段飞轮。Python 已挂 Admin：

| 层 | Collection | Admin | payload |
|----|------------|-------|---------|
| OP | `feeding_op_cues` | `/v1/admin/op-cache` list/bulk/delete | `op_rail` |
| 事件 | `feeding_intents` | intent-cache（Hub 已有） | 事件短语 |
| 数量 | `feeding_qty_cues` | `/v1/admin/qty-cache` list/bulk/delete | `quantity` |
| R·source | `feeding_read_source_cues` | `/v1/admin/read-source-cache` list/bulk/delete | `read_source` |
| R·record 播报 | `feeding_history_speak_cues` | history-speak-cache（Hub 已有） | `point_pick`/`open_speak` |
| 旧 TYPE | `feeding_type_cues` | type-cache（热路径已停） | 遗留 |

规则缺口仍 TIME/QTY → JSONL；无时间飞轮。R·predict / U·D 定位尚无独立 cue admin。

Go Hub 现状：意图智能体门户四卡（事件 / 类型 / 查史播报 / 规则缺口）；类型页仍教灌旧 `target_type`；缺 OP/数量/R·source。

约束：Admin JWT + gateway 注入 `X-Admin-Password`；不直连 Chroma；Admin 不碰 usage；`/voice/admin/api/*` 反代已覆盖；不新增 Redis / 背景循环。

## Goals / Non-Goals

**Goals:**

- Hub 可整理热路径三层新飞轮：OP、数量、R·数据源（list/bulk/delete）。
- 门户按运维日常重排：飞轮卡 + 规则缺口；去掉类型向量主卡。
- 类型向量深链保留并标明遗留；规则缺口文案区分 TIME 主工单与 QTY+飞轮。

**Non-Goals:**

- R·predict / locate 独立飞轮 Hub（等 Python admin）。
- 为新三层补 dedupe（上游无此路由）。
- 改 Python 飞轮/写路径或拆 Go batch API。
- 合并多飞轮为单页 Tab；新建 MySQL 表；Hub 一键 purge JSONL。

## Decisions

### D1：三组代理镜像 type-vectors，但无 dedupe

Hub 路径（`*-vectors` 与既有对称）→ Python `*-cache`：

| Hub | Python |
|-----|--------|
| `GET/POST/DELETE /voice/admin/api/op-vectors*` | `/v1/admin/op-cache*` |
| `…/qty-vectors*` | `/v1/admin/qty-cache*` |
| `…/read-source-vectors*` | `/v1/admin/read-source-cache*` |

每组：`List` / `BulkUpsert` / `Delete`；`PythonAIClient` + `VoiceAdmin*Ctrl` + `VerifyVoiceAdminPassword`；`api/v1` DTO。

**备选**：gateway 直打 Python → 否决（破坏 Hub→Go 与口令惯例）。

### D2：三张独立静态页，克隆 history-speak / type 骨架

- `op-vector-admin.html`：提示整段 document；payload `{op_rail}`（create|read|update|delete|exit|other）；**无**「整理重复」。
- `qty-vector-admin.html`：payload `{quantity}` 正整数；无量勿灌。
- `read-source-vector-admin.html`：payload `{read_source: record|predict}`。

登记：`admin-modules.js`（`showInNav: false`，经门户进入）、`admin_static_pages.go`、`gateway_app_auth_exempt.go`。

### D3：门户重排（方案 A：改卡，非下拉）

`intent-agent-admin.html` 卡片：

| 区 | 标题 | 目标 |
|----|------|------|
| 飞轮 | OP 轨 | op-vector-admin.html |
| 飞轮 | 事件 | intent-vector-admin.html |
| 飞轮 | 数量 | qty-vector-admin.html |
| 飞轮 | R·数据源 | read-source-vector-admin.html |
| 飞轮 | 查史播报 | history-speak-vector-admin.html |
| 工单 | 规则缺口 | rule-gaps-admin.html |

**不**再放「类型向量」主卡。`type-vector-admin` 模块保持 `showInNav: false`；页头加遗留横幅：「热路径已停，新种子请用 OP 轨」。

### D4：规则缺口文案微调

- TIME：主工单——扩本地规则须改 Python 发版。
- QTY：可进箱；优先数量飞轮整段灌种，减少重复 QTY LLM。
- 默认筛选语义不变（空 dimension → 不传参，吃 Python TIME+QTY）。

### D5：部署边界

- **voice-service**：三组代理。
- **gateway-app**：静态页 + 门户/类型页文案 + 模块登记。
- **python-ai-talk**：须已含 op/qty/read-source-cache（同发或先发）。

## Risks / Trade-offs

- [Python admin 未发版] → 新页 5xx → 页内提示检查 `PYTHON_AI_TALK_URL` / 兄弟仓；门户其它卡不受影响。
- [运维仍灌类型向量] → 门户去卡 + 类型页横幅；深链不删以免书签断裂。
- [误灌短词 document] → 页 hint 强调「整段原文」；阈值由 Python 保证。
- [无 dedupe] → 运维靠 list+delete；接受与 intent 页按钮不对称。

## Migration Plan

1. 确认/发版 Python op/qty/read-source-cache。
2. 发版 voice（代理）+ gateway-app（三页/门户/类型横幅/缺口文案）。
3. 回滚：撤三页与路由、恢复门户四卡；Chroma 数据保留。

## Open Questions

- （无阻塞。predict/locate admin 落地后另开 Hub change。）
