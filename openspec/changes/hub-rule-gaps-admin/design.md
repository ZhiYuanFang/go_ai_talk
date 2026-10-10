## Context

Python `intent-segment-multidim` 已交付规则缺口收件箱：

- 存储：`{chroma_persist_dir}/../rule_gaps/gaps.jsonl`（非 MySQL、非向量库）
- API：`GET /v1/admin/rule-gaps`、`PATCH /v1/admin/rule-gaps/{gap_id}`（status=open|dismissed）
- 行字段：`id`、`dimension`、`segment_text`、`utterance`、`reason`、`rule_result`、`llm_used`、`llm_result`、`device_no`、`status`、`created_at`

Go 侧「意图向量缓存」已建立范式：Hub 页 → `/voice/admin/api/intent-vectors*` → `PythonAIClient` → `/v1/admin/intent-cache*`。rule-gaps 的 Go Hub tasks（Python 仓 5.1–5.3）未做。

约束：voice Admin 口令；跨服务只经 HTTP；不新增 Redis；不新增背景循环；Admin 非 App 不碰 usage。

## Goals / Non-Goals

**Goals:**

- Hub 可分页浏览规则缺口（默认 open），按 dimension/status 筛选。
- 展示整句、段、reason、LLM 摘要等，支持标 dismissed。
- 导航与「意图向量缓存」并列，文案说明人工改 Python 规则发版、非热更新。

**Non-Goals:**

- 不在 Go 建表或同步 JSONL。
- 不改 Python 写入逻辑 / 去重窗 / 缺口语义。
- 不自动补规则、不写 `feeding_intents`。
- 不提供从 Hub 热改规则代码。

## Decisions

### 1. 纯代理，对齐 intent-vectors

- **选择**：`PythonAIClient` 增加 `ListRuleGaps` / `PatchRuleGapStatus`；控制器 `VoiceAdminRuleGapsCtrl`；路径建议：
  - `GET /voice/admin/api/rule-gaps`
  - `PATCH /voice/admin/api/rule-gaps/{id}`
- 鉴权：`VerifyVoiceAdminPassword`（同 intent-vectors）。
- **理由**：零新存储、契约已存在；运维路径统一在 voice Admin。
- **备选**：gateway 直反代 Python → 破坏「Hub 只打 Go」与口令注入惯例，否决。

### 2. 响应透传 Python 字段

- **选择**：list 返回 `total/offset/limit/items`；item 字段名与 Python JSON 对齐（snake 或在 DTO 用 json tag 映射为 camel——实现时与 intent-vectors 透传风格一致；若 Python 为 snake_case，Hub 页直接读 snake 或 Go 映射 camel，**优先与 intent-vectors 页面习惯一致**，建议 Go DTO 用 camel `json` 映射）。
- **理由**：少一层变换错误；页面列：dimension、segment、utterance、reason、llm、device、created、status、操作。

### 3. Hub 静态页

- **选择**：`resource/public/rule-gaps-admin.html`；`admin-modules.js` 标题「规则缺口」；`admin_static_pages.go` + `gateway_app_auth_exempt.go`（若静态页需与 intent-vector 同级登记）。
- 筛选项：dimension（空=全部）、status（默认 open，可选 dismissed/空=全部若 Python 支持 status="" 不过滤——Python `status` 空串不过滤）。
- 分页：offset/limit（默认 100）。
- 行操作：标已处理 → PATCH `dismissed` → 刷新。

### 4. 部署依赖

- **选择**：依赖运行中的 `python-ai-talk` 与 volume 上已有 `gaps.jsonl`（可无文件=空列表）。
- 页脚/hint：说明「数据在 Python 卷 `data/rule_gaps/gaps.jsonl`；整理规则须改 Python 发版」。

## Risks / Trade-offs

- [Risk] Python 未部署或 volume 未挂载 → 列表失败 → **Mitigation**：错误文案提示检查 `PYTHON_AI_TALK_URL` / python 服务。
- [Risk] JSONL 全量读在 Python 侧随文件变大变慢 → **Mitigation**：首版沿用 Python 实现；量大再兄弟仓优化。
- [Risk] 运维误以为 Hub 可改规则 → **Mitigation**：页面 hint 明确「仅收件箱，改规则须发版」。

## Migration Plan

- 发版 gateway-app（静态页/导航）+ voice-service（代理 API）；Python 无需改。
- 回滚：去掉页与路由即可；JSONL 数据保留。

## Open Questions

- （无阻塞。若产品只要只读列表、不要 dismissed，仍保留 PATCH 以对齐 Python API。）
