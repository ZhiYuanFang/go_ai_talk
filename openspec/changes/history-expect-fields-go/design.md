## Context

Flutter 在 add 时提交 `expectStartTime` / `expectEndTime` / `expectAmount`。本仓负责 History 域持久化与回传。产品：非必填；update/stop 冻结；服务端不补算。AGENTS 有「既有 API 不可改结构须升 v+1」；本变更默认 **只增可选字段**。

## Goals / Non-Goals

**Goals:**

- 可空三列；add 写入；list/WS/缓存读出。
- update / end-latest **UPDATE 不写** expect 列，避免 `int64` 零值抹库。
- `expectAmount` 与预测/Flutter 一致为浮点。
- 最少 DB 操作；无新 ticker。

**Non-Goals:**

- Flutter 打戳、准确度 UI。
- 服务端填 expect*。
- 新建微服务。

## Decisions

1. **API 版本策略**  
   现有 App add/list/update **同一 path** 增加 `omitempty` 可选字段，**不**新建 v2。  
   *若评审驳回：再开 v2 + Flutter 切路径。*

2. **存储与类型**  
   - `expect_start_time` / `expect_end_time`：可空 BIGINT，unix 秒。  
   - `expect_amount`：可空 **DOUBLE/DECIMAL**（实现选 DOUBLE 即可），**禁止** INT。  
   - 实体：`*int64` / `*float64`（或 Null*），禁止裸 `int64` 把「未传」与 0 混用。

3. **Insert vs Update 写列**  
   - `historyInsertData`：有非 nil 才写入对应列；全缺则不出现在 INSERT。  
   - `historyUpdateData`：**永远不包含**三 expect 列。  
   - `EndLatestDeviceHistoryIfMatch` 继续只改 endTime/remark（已满足冻结）。

4. **Update 冻结**  
   即使 Req JSON 带 expect*，handler MUST 忽略；不靠「先读再合并」，靠「不写列」。

5. **读路径齐全**  
   list select fields、`historyRowToEntity`、`historyToNotifyPayload`、cache patch、remote HTTP JSON map 均带 expect*；admin add 若落同表则同 INSERT 规则。

6. **WS**  
   与 list 同一序列化。

7. **DB 连接**  
   仅 `history-service` + `HISTORY_DB_LINK` / `database.default`。

## Risks / Trade-offs

- [AGENTS 升版字面冲突] → 默认增量可选；评审可改 v2。  
- [Flutter 先发] → 建议 Go 先或同发。  
- [误把 expect 写入 Update map] → 零值抹库；tasks 明确禁写 + 评审 grep。  
- [旧行 NULL] → 准确度冷启动，已接受。

## Migration Plan

1. migration + history-service。  
2. 再发 / 同发 Flutter twin。  
3. 回滚：停写列；读端 omitempty。

## Open Questions

- 无（升版交评审；可空类型 / 禁写列 / amount=float 已定）。
