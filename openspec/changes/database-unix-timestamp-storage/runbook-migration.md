# Runbook：时间列迁移（Unix 秒）

## 发布顺序（建议）

1. 停写或缩窗：降低写入并发（可选，视表大小）。
2. DDL：加新列 `*_unix_sec BIGINT NULL` 或 `MODIFY`（仅低峰允许原地改类型时）。
3. 回填：`UPDATE ... SET col_sec = <解析旧值>`；批处理 + 校验行数。
4. 应用发版：双读（先读新列，NULL 则读旧）→ 切换只读新列。
5. 观察期后：`DROP` 旧列或保留只读备份。

## 回滚

- 保留旧列至稳定期；应用配置开关将读指回旧列。
- 回填错误时：从新列复制回旧列仅在旧列仍可写且类型兼容时可行，须事先演练。

## 校验 SQL（示例）

```sql
-- 行数一致（迁移前后）
SELECT COUNT(*) FROM your_table;
-- 秒值合理（约在 2000～2100 年对应区间外则异常）
SELECT MIN(col_sec), MAX(col_sec) FROM your_table WHERE col_sec IS NOT NULL;
```

## 与现网 DDL 参考

- 见仓库 `hack/ddl_wx_and_app_version.sql` 中 `release_date` bigint 秒示例。

## 未完成项（须在有 DB 的环境执行）

- 在**测试库**执行完整迁移与回填（对应 tasks.md 2.2）。
- 全量 API `string` → `int64` 与 history/device 服务读写改造（对应 tasks.md 3.x）须在 DDL 完成后串联发布。
