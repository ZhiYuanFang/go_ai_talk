## Context

`predict-pro-idea-hub` 落地后，submit 成功且库中有行，但 App `list` 与 Hub admin 列表均为空。查询统一使用 GoFrame `WhereNull("deleted_at")`（`deleted_at IS NULL`）。线上确认新增行 `deleted_at` 默认值为 `0`（表结构/默认值与 Ensure DDL 中 `NULL DEFAULT NULL` 不一致，且 Ensure 在表已存在时跳过 CREATE）。本仓库 UCG 等域习惯以 `deleted_at = 0` 表示未删。

## Goals / Non-Goals

**Goals:**

- App / Admin 列表能返回 `deleted_at = 0`（及存量 `NULL`）的活跃想法。
- 按 id 加载、改展示文案、软删前置校验与「未删」语义一致。
- Ensure/DDL 与线上默认 `0` 对齐，避免新环境再分裂。
- 改动局限在 `device-service` 想法模块；HTTP 契约形状不变。

**Non-Goals:**

- 不改日限额算法、不改 Pro/JWT 门禁、不改 Flutter/Hub 前端。
- 不做跨库迁移工具；不引入 Redis。
- 不顺带改其它域表的软删约定。

## Decisions

1. **未删谓词：`deleted_at IS NULL OR deleted_at = 0`**
   - 理由：兼容线上默认 `0` 与 DDL 曾写的 `NULL`；与 UCG「0=未删」一致。
   - 备选 A：仅改成 `Where("deleted_at", 0)` —— 会漏掉存量 `NULL`。
   - 备选 B：只洗数把 `0` 改成 `NULL`、查询仍 `WhereNull` —— 治标不治本，Ensure 与插入默认仍可能再写 `0`。
   - 选择：查询侧兼容两者；软删仍写入 `time.Now().Unix()`（`>0`）。

2. **集中辅助函数，避免漏改**
   - 在 `predict_pro_idea.go` 提供 `wherePredictProIdeaAlive(m)`（或等价 Where 片段），供 list App/Admin、`loadActiveIdea`、Update display、SoftDelete 共用。
   - 日限额 COUNT **故意不加** alive 过滤（规格已要求含软删）。

3. **EnsureSchema / DDL**
   - 新建表：`deleted_at BIGINT NOT NULL DEFAULT 0`，注释写明 `0=未删；>0=软删 unix 秒`。
   - 表已存在：不强制 ALTER（避免锁表风险）；靠查询兼容即可。可选：若需文档一致，可幂等 `MODIFY` 仅在确认低风险时执行——**默认不做 ALTER**，以代码兼容为主。
   - Insert 可不显式写 `deleted_at`（依赖默认 `0`）；若担心其它环境 DEFAULT NULL，Insert 可显式 `"deleted_at": 0` 更稳。

4. **无数据洗数**
   - 修复部署后 `deleted_at=0` 行立即可见；已软删（`>0`）仍隐藏。

## Risks / Trade-offs

- [Risk] 某环境若误用非 0 哨兵表示未删 → Mitigation：仅承认 0/NULL；规格写死。
- [Risk] 忘记改某一处 `WhereNull` → Mitigation：统一 alive helper + 任务清单全覆盖。
- [Trade-off] 双值兼容略宽于「只认 0」，换取零洗数与平滑过渡。

## Migration Plan

1. 合并并部署 **device-service**（含 Ensure 与查询修复）。
2. 验证：submit 后 list 非空；Hub 可见；软删后 list 消失；日限额仍计软删。
3. Rollback：回滚二进制即可；数据无需回滚。

## Open Questions

（无；线上已确认新增 `deleted_at` 默认为 0。）
