## Context

意图门户链到 9 个 `*-vector-admin.html`。右侧已是批量 JSON 上传。Python（`python_ai_talk`）各 `*_cue_store.py` / `intent_cache_store.py` 定义 payload。预测 Hub 示例仍是两字段；Python `as_predict_payload` 现要求 `predict_channel`、`time_answer`、`predict_aspect`。`feeding_type_cues` 有页但门户不链。

## Goals / Non-Goals

**Goals:**

- 每维一页一段可复制提示词，交给其它 LLM 后产出可直接粘贴的种子数组。
- 预测示例与门户说明补上 `predict_aspect`。

**Non-Goals:**

- 不调 LLM、不把提示词存库。
- 不把类型向量重新放进门户。
- 不改 bulk API 契约。

## Decisions

1. **提示词写在各 HTML 的 `<textarea readonly>`，旁有「复制」按钮**
   - 放在该页右侧批量区标题下、JSON 输入框上。
   - 文案含：角色一句、输出仅为 JSON 数组、document 规则、payload 字段与枚举、禁止项（确认词当 document；时间维绝对 Unix；定位维 historyId）。
   - **备选**：门户一页集中全部提示词 — 字段一变要跳页对照，否决；以各维页为准，门户只改预测卡片一句。

2. **字段以当前 Python 为准（摘录进静态文案）**
   - OP `op_rail`: create|read|update|delete|exit|other
   - 事件：短说法；`event_id` / `event_name` / `kind`(parent|leaf)
   - 数量：`quantity` 整数
   - 时间：`kind` instant|interval|duration_to_now|open_start；点 `day` today|yesterday|day_before，`hour`/`minute` 或 `period` morning|noon|afternoon|evening|night；`duration_to_now` 用 `duration_minutes`
   - 备注：`remark` 可空串
   - 数据源：`read_source` record|predict
   - 预测：`predict_channel` active|next；`predict_aspect` when|amount|duration；`time_answer` 仅 when 时 clock|remain，否则空串
   - 查史：`point_pick` latest|prefer_completed；`open_speak` elapsed|in_progress
   - 定位：`strategy` latest|open_timer|prefer_completed

3. **类型页**
   - 提示词首行声明热路径已停。不加入 `intent-agent-admin.html` 门户。

## Risks / Trade-offs

- [Python 枚举以后再变，静态文案过期] → 提示词注释写明对照仓库与文件名，变更时人工改 HTML。
- [复制失败] → `navigator.clipboard`，失败时选中文本。

## Migration Plan

- 静态发布后 Ctrl+F5。无数据迁移。

## Open Questions

- （无。）
