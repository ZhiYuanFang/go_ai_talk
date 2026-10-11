## Context

Hub 顶栏三个平级模块（意图向量 / 类型向量 / 规则缺口）造成认知噪音。Python 意图智能体向量面：

| 集合 | Admin 上游 | Hub |
|------|------------|-----|
| `feeding_intents` | `/v1/admin/intent-cache*` | 已有 |
| `feeding_type_cues` | `/v1/admin/type-cache*` | 已有 |
| `feeding_history_speak_cues` | 计划 `/v1/admin/history-speak-cache*`（store 已有，路由待挂） | 无 |
| rule_gaps（非向量） | `/v1/admin/rule-gaps*` | 已有 |

TIME/QTY 无独立向量库，不进门户「向量」卡。

约束：Admin JWT + gateway 注入 `X-Admin-Password`；不直连 Chroma；Admin 不碰 usage。

## Goals / Non-Goals

**Goals:**

- 顶栏唯一入口「意图智能体」→ 门户页。
- 门户聚合四子能力：事件向量、类型向量、查史播报向量、规则缺口。
- 查史播报：Go 代理 + 独立整理页（镜像 type-vectors）。
- 旧深链 URL 仍可用。

**Non-Goals:**

- 合并三页为单页 Tab（首版门户+独立页）。
- 改 intent/type/rule-gaps 代理逻辑或 Python 飞轮热路径。
- 为 TIME/QTY 新建向量 Admin。
- 本仓实现 Python admin 路由（兄弟仓职责；契约对齐即可）。

## Decisions

### D1：门户页（方案 A），非顶栏下拉

- 新页：`/device/admin/intent-agent-admin.html`（`resource/public/intent-agent-admin.html`）。
- `admin-modules.js`：新增 `intent-agent-admin`，`title: 意图智能体`，`showInNav: true`。
- 原 `intent-vector-admin` / `type-vector-admin` / `rule-gaps-admin`：`showInNav: false`（保留模块登记供深链/文档）。
- 门户四卡文案与链接：

  | 门户标题 | 目标 |
  |----------|------|
  | 事件向量 | intent-vector-admin.html |
  | 类型向量 | type-vector-admin.html |
  | 查史播报向量 | history-speak-vector-admin.html |
  | 规则缺口 | rule-gaps-admin.html |

- 子页头部链回门户（最小改：副标题或「返回意图智能体」）。

**备选 B（下拉）**：需改 `admin.html` 渲染嵌套模块，收益小，否决。

### D2：查史播报代理镜像 type-vectors

上游（Python，与 type-cache 同形）：

- `GET /v1/admin/history-speak-cache?offset&limit`
- `POST /v1/admin/history-speak-cache/bulk`（items: `{document, payload}`；payload=`point_pick`|`open_speak`）
- `POST /v1/admin/history-speak-cache/dedupe`
- `DELETE /v1/admin/history-speak-cache/{id}`

Go Hub：

- `GET/POST/DELETE /voice/admin/api/history-speak-vectors*`（命名用 vectors，与 intent/type 对称）
- `PythonAIClient`：`ListHistorySpeakCache` / `BulkUpsert` / `Dedupe` / `Delete`
- `VoiceAdminHistorySpeakVectorsCtrl` + `VerifyVoiceAdminPassword`
- 静态页：`history-speak-vector-admin.html`（list/bulk/dedupe/delete；示例 payload）

若 Python 路由未上线：Hub 列表失败 → 页内提示检查 `PYTHON_AI_TALK_URL` / 兄弟仓发版。

### D3：登记与鉴权

- `admin_static_pages.go` + `gateway_app_auth_exempt.go`：门户 + 查史播报页（同 intent-vector 模式，壳页可匿名加载，API 仍须 Admin JWT）。
- 无需新 proxy 前缀；无新 env。

### D4：部署边界

- **gateway-app**：静态 + 导航。
- **voice-service**：history-speak 代理。
- **python-ai-talk**：history-speak-cache 路由须同发或先发。

## Risks / Trade-offs

- [Python admin 未挂] → Hub 查史页 5xx/空 → 提示兄弟仓发版；门户其它三卡不受影响。
- [运维书签旧顶栏三项] → 深链仍可用；顶栏只见门户。
- [命名「事件向量」vs 文件名 intent-vector] → 门户文案纠正即可，不强制改文件名。

## Migration Plan

1. 确认/发版 Python `history-speak-cache` admin。
2. 发版 voice（代理）+ gateway-app（门户/导航/新页）。
3. 回滚：恢复三模块 `showInNav`、撤门户与查史页/路由；Chroma 数据保留。

## Open Questions

- （无阻塞。Python 路由路径以 `/v1/admin/history-speak-cache` 为准，若兄弟仓微调前缀，apply 时对齐实际路由。）
