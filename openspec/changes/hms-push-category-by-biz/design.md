## Context

`HmsSender.Send` 已从 `payload.Data["bizType"]` 读取业务类型，用于 `click_action` intent。`android.notification` 目前含 `click_action`、`priority`、`importance`（及可选 badge），缺少华为消息自分类字段 `category`。

业务侧已稳定使用：
- `predict_imminent`：预测临近可见提醒（voice）
- `ucg_alert`：UCG 可见内容推送
- `ucg_silent_badge`：UCG 静默角标

华为云通知分类枚举中，`WORK` = 工作提醒，`IM` = 即时通讯；值须大写。

## Goals / Non-Goals

**Goals:**

- 在 HMS 可见路径按 `bizType` 写入正确 `category`，提升服务与通讯类送达效果。
- 明确省略规则：静默角标与未知类型不误标。

**Non-Goals:**

- 不改 APNs / MiPush。
- 不扩展 `bizType` 常量集合、不改 internal 推送 HTTP 契约。
- 不在本变更内办理华为自分类资质申请。
- 不新增测试文件（仓库现阶段约定）。

## Decisions

1. **映射表（仅 HMS）**
   - `predict_imminent` → `WORK`
   - `ucg_alert` → `IM`
   - `ucg_silent_badge` → **省略** `category` 键
   - 空 / 其它 `bizType` → **省略** `category` 键  
   **理由**：与产品语义对齐；静默与未知类型不猜分类，避免违规或被厂商纠偏。  
   **备选**：静默也标 `IM` —— 已否决（产品确认不传）。

2. **实现位置**  
   在 `push_hms.go` 构建 `androidNotif` 时，根据已解析的 `bizType` 条件赋值（小函数或 switch 即可）。  
   **理由**：分类是厂商载荷细节，不泄漏到 dispatcher / 调用方。  
   **备选**：调用方在 `data` 里带 `category` —— 拒绝，避免各域重复且易写错大小写。

3. **枚举大小写**  
   固定大写 `WORK` / `IM`（华为文档）。  
   **理由**：小写可能导致无效或被拒。

4. **与顶层 notification 的关系**  
   `category` 写在 `message.android.notification`，与现有可见消息（顶层 `message.notification` + android `click_action`）并存；静默无顶层 notification 时本就不写 `category`（因 `ucg_silent_badge` 省略）。

## Risks / Trade-offs

- [华为未授权自分类，`WORK`/`IM` 被拒或降级] → 运维确认资质；失败走现有 `send_failed` 日志；可回滚省略 `category`。
- [未来新增 bizType 忘记映射] → 默认省略（安全）；新增类型时在本 capability 增场景。
- [与 `channelId` 注释掉的历史问题混淆] → `category` 是云通知分类，不是本地 channelId；代码注释标明。

## Migration Plan

1. 合并后仅滚动 `push-service`。
2. 用 `predict_imminent` / `ucg_alert` 各发一条 HMS，抓请求体确认 `category`；`ucg_silent_badge` 确认无该字段。
3. 回滚：回退 commit 并重启 push-service（无 DB/配置迁移）。

## Open Questions

- （无）产品已确认：静默与未知不传；预测 `WORK`、UCG 可见 `IM`。
