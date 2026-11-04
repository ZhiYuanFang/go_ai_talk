## Context

Flutter 将在 add 时提交 `expectStartTime` / `expectEndTime` / `expectAmount`。本仓负责 History 域持久化与回传。产品：字段非必填；update/stop 冻结 expect*；服务端不强制补算。AGENTS 有「既有 API 版本不可改结构须升 v+1」条文；本变更语义为 **只增可选字段、不改既有必填/含义**，与「老客户端忽略未知键、不传仍成功」一致。

## Goals / Non-Goals

**Goals:**

- DB/实体可选三字段；add 写入；list/WS 读出。
- update 路径不覆盖已有 expect*。
- 最少 DB 操作；无新后台 ticker。

**Non-Goals:**

- Flutter 打戳、准确度 UI。
- 服务端根据算法填 expect*。
- 新建微服务进程（无 CI/ACR 扩容项）。

## Decisions

1. **API 版本策略**  
   在现有 App 历史 add/list/update **同一 path** 上为 Req/Res 增加 `omitempty` 可选字段，**不**新建 v2 path。  
   *理由：不破坏既有字段；老客户端不传/忽略；产品明确非必填。*  
   *若评审驳回：再开 DeviceHistoryEventAdd v2 + list 实体升版，并由 Flutter twin 切路径。*

2. **存储**  
   History 表新增可空列（bigint 秒 / 用量数值），或与现 raw 存储方式一致的可空字段；add INSERT 带入；update 默认 UPDATE 语句 **不包含** expect 列。

3. **Update 冻结**  
   即使请求 JSON 带 expect*，update handler MUST 忽略，保留库中原值。

4. **WS**  
   realtime 推送 payload 使用与 list 相同的实体序列化，含 expect*。

5. **DB 连接**  
   仅 `history-service` / `HISTORY_DB_LINK` 域内表；不新增跨库。若 tasks 含 migration，注明进程与配置组 `default`（history）。

## Risks / Trade-offs

- [AGENTS 升版字面冲突] → design 已选增量可选；评审可改 v2。  
- [Flutter 先发、Go 未发] → 未知字段可能被框架丢弃或 4xx → 建议 Go 先或同发。  
- [旧行 null] → 准确度冷启动，产品已接受。

## Migration Plan

1. migration + history-service 发版。  
2. 再发 / 同发 Flutter twin。  
3. 回滚：停写列；读端 omitempty 兼容。

## Open Questions

- 无（升版争议交评审，默认增量可选）。
