## 1. 查询与软删语义

- [x] 1.1 在 `predict_pro_idea.go` 增加未删过滤辅助（`deleted_at IS NULL OR deleted_at = 0`），替换所有 `WhereNull("deleted_at")`（App list、Admin list/Count、`loadActiveIdea`、Update display、SoftDelete）
- [x] 1.2 确认日限额 `countPredictProIdeaToday` **不加** alive 过滤（含软删仍计数）
- [x] 1.3 SoftDelete 仍写入 `deleted_at = unix`；Insert 可选显式 `"deleted_at": 0` 以防环境 DEFAULT 不一致

## 2. Schema 对齐

- [x] 2.1 更新 `schema_predict_pro_idea.go` 新建表 DDL：`deleted_at BIGINT NOT NULL DEFAULT 0` 及中文注释（0=未删；>0=软删）
- [x] 2.2 Ensure 对已存在表不做强制 ALTER（靠查询兼容）

## 3. 校验与部署说明

- [x] 3.1 本地/目标环境确认：`deleted_at=0` 行出现在 App list 与 Hub；软删后消失
- [x] 3.2 回复中列出重部署：`device-service`；不必重部署 gateway（路径/鉴权未变）
