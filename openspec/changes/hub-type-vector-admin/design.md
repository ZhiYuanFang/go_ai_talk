## Context

兄弟仓 `type-keyword-flywheel` 已交付：

- Chroma 集合 `feeding_type_cues` + Admin API：`GET/POST/DELETE /v1/admin/type-cache*`（list / bulk / dedupe / delete）
- 规则缺口写入收窄为 TIME/QTY；`GET /v1/admin/rule-gaps` 在 `dimension` 省略时默认只合并 `time` + `quantity`
- 明确兄弟仓任务：Go Hub 类型向量页 + 反代；规则缺口页默认 TIME/QTY；可选清理存量 JSONL

Go 侧已有：

- 「意图向量缓存」：`intent-vector-admin.html` → `/voice/admin/api/intent-vectors*` → `PythonAIClient` → `/v1/admin/intent-cache*`
- 「规则缺口」：`rule-gaps-admin.html` + 代理已就绪，但 dimension 控件 placeholder 仍为「空=全部，如 time/event」

约束：voice Admin 口令；跨服务只经 HTTP；不新增 Redis；不新增背景循环；Admin 非 App 不碰 usage；Go MUST NOT 直读 Python JSONL。

## Goals / Non-Goals

**Goals:**

- Hub 可 list/bulk/dedupe/delete 类型关键字向量（payload：`target_type` / `op` / `history_mode`）。
- 导航与「意图向量缓存」并列，独立静态页。
- 规则缺口页默认语义与 Python 对齐：空 dimension = TIME+QTY；文案不再暗示全维工单箱。
- 文档/任务注明：存量 legacy 行可选由运维在 Python 卷 filter-rewrite 或清空。

**Non-Goals:**

- 不在 Go 建类型向量表或同步 Chroma。
- 不改 Python 飞轮写路径 / type-cache / rule_gaps 服务端逻辑。
- 不提供 Hub「一键清理 JSONL」按钮或 Go 侧 purge API。
- 不合并意图向量与类型向量为单页 Tab（首版独立页，与现有模块一致）。

## Decisions

### 1. 类型向量：纯代理，镜像 intent-vectors

- **选择**：`PythonAIClient` 增加 `ListTypeCache` / `BulkUpsertTypeCache` / `DedupeTypeCache` / `DeleteTypeCache`；控制器建议 `VoiceAdminTypeVectorsCtrl`；路径：
  - `GET /voice/admin/api/type-vectors`
  - `POST /voice/admin/api/type-vectors/bulk`
  - `POST /voice/admin/api/type-vectors/dedupe`
  - `DELETE /voice/admin/api/type-vectors/{id}`
- 上游：`/v1/admin/type-cache*`（命名保留 Python cache；Hub URL 用 vectors 与意图侧对称）。
- 鉴权：`VerifyVoiceAdminPassword`。
- **理由**：零新存储、契约已存在；运维路径统一。
- **备选**：gateway 直反代 Python → 破坏「Hub 只打 Go」与口令注入惯例，否决。

### 2. 类型向量页：独立 HTML，克隆意图页骨架

- **选择**：`resource/public/type-vector-admin.html`；`admin-modules.js` 标题「类型向量」或「类型向量缓存」；登记 `admin_static_pages.go` + `gateway_app_auth_exempt.go`。
- Payload 提示与 bulk 示例使用 `{target_type, op, history_mode}`（如 `history`/`read`/`point`），勿复用事件 phrases 字段。
- DTO/JSON 风格与 intent-vectors 透传习惯一致。

### 3. 规则缺口 UX：依赖 Python 默认 + 改 Hub 文案/控件

- **选择**：dimension 空时**不传** `dimension` 查询参数（继续吃 Python 默认 TIME+QTY）；更新 placeholder/hint 为「默认 TIME+QTY；可填 time / quantity；排障可填 legacy 维度」。
- 可选：下拉 `默认(TIME+QTY) | time | quantity`，保留「自定义」文本框以便 `dimension=type` 排障。
- **不**在 Go 控制器硬编码过滤 type/event/shape（权威在 Python）。
- **理由**：避免双源过滤语义漂移；最小改动收口运维认知。

### 4. 存量 JSONL 清理：运维可选，非本仓代码

- **选择**：tasks/design 写明 runbook——在 Python 数据卷对 `data/rule_gaps/gaps.jsonl` 做 truncate 或只保留 `dimension in (time, quantity)` 的 filter-rewrite；本变更不实现 purge API。
- **理由**：Python 无 delete/purge；Go 禁止碰 JSONL；「可选」与兄弟仓 task 3.4 措辞一致。

## Risks / Trade-offs

- [Risk] Python type-cache 未发版 → Hub 列表失败 → **Mitigation**：错误提示检查 `PYTHON_AI_TALK_URL` / 兄弟仓发版。
- [Risk] 运维仍用自定义 dimension 看 legacy 噪音 → **Mitigation**：默认路径不传 dimension；hint 标明 legacy 仅排障。
- [Risk] 未清 JSONL 时文件仍大、Python 全量读慢 → **Mitigation**：运维可选清理；性能仍属 Python 侧后续优化。
- [Risk] bulk 误写错误 target_type → **Mitigation**：页内示例与字段说明；同 document 跳过策略由 Python 保证。

## Migration Plan

1. 发版 voice-service（type-vectors 代理）+ gateway-app（静态页/导航/规则缺口文案）。
2. 确认 python-ai-talk 已含 type-cache 与 rule_gaps 默认过滤。
3. （可选）运维清理存量 gaps.jsonl。
4. 回滚：去掉类型向量页与路由；规则缺口文案回退不影响 API；Chroma/JSONL 数据保留。

## Open Questions

- （无阻塞。若后续产品坚持 Hub 一键 purge，须另开兄弟仓 API + 新 change。）
