## Context

`HmsSender.Send` 当前把 `title`/`body` 仅放在 `message.android.notification`，并始终带 `android.data`；`click_action` 固定 `type:1` 且无 `intent`。华为控制台成功样例通常含顶层 `message.notification`，`android.notification.click_action` 为 `type:3`（打开应用）或 `type:1`+完整 intent。现象：控制台推送客户端可见，服务端 `send_ok` 后不可见。另：HTTP 2xx 即视为成功，未校验响应 JSON `code==80000000`，可能假成功。

约束：仅改 HMS 发送路径；中文注释；不新增测试文件；prod 下失败须 Warning 可观测。

## Goals / Non-Goals

**Goals:**

- 非静默+有 alert：发出可被系统通知中心展示的 HMS **通知消息**（对齐控制台主流形态）。
- 成功判定以业务 `code` 为准；失败可观测。
- 保留自定义 data（biz 字段），不阻碍通知展示。

**Non-Goals:**

- 不改 APNs/MiPush。
- 不改 push 注册/token 表、不改 voice predict 业务。
- 不实现完整「点击深链到某页面」产品（默认打开应用即可；intent 可选 env）。
- 不解析华为送达回执/回执订阅。

## Decisions

### D1. 非静默可见推送载荷

```json
{
  "message": {
    "notification": { "title": "胖宝", "body": "<alert>" },
    "android": {
      "notification": {
        "click_action": { "type": 3 },
        "badge": { "add_num": 0, "set_num": <badge>, "class": "com.fzy.pangbao.MainActivity" }
      },
      "data": "<json string of biz fields>"
    },
    "token": ["..."]
  }
}
```

- 顶层 `notification` 负责托盘标题正文。
- `android.notification` 只放 Android 专有项（click_action、badge）；避免再重复依赖仅 android 内 title/body。
- 若 `PUSH_HMS_CLICK_INTENT` 非空：改用 `click_action.type=1` + `intent=<env>`。
- 备选（否）：只修 intent、保留现状结构——仍缺顶层 notification，与文档样例差距大。

### D2. 静默 / 无 alert

- `silent==true` 或 alert 空：不发「完整可见通知」；可仅 badge / data（保持现有静默语义），不得伪造空 body 通知骚扰。
- UCG silent badge 等调用方行为不变。

### D3. `data` 内容

- 继续用现有 `buildHmsMessage`（badge/silent/alert 及可扩展）；放在 `android.data`（覆盖 `message.data` 的官方语义下二选一即可，本设计固定 `android.data`）。
- 不把完整 token 写入 data。

### D4. 成功判定

- HTTP 非 2xx → 失败（现有）。
- HTTP 2xx 时解析 JSON：`code` 为 `80000000`（字符串或数字）才成功；否则 `send_failed`，Warning 含 `code` 与截断 msg（无 token）。
- 无效 token 类 code/msg 仍走 `invalidToken=true` 删除逻辑。

### D5. 标题

- 默认 title `"胖宝"`（与现网一致）；暂不引入可配置 title env（需要时再加）。

## Risks / Trade-offs

- [部分机型对「notification+data」组合前台行为不同] → 以通知展示为先；与控制台一致优先。
- [badge.class 包名/Activity 错误] → 沿用现网 `com.fzy.pangbao.MainActivity`；若角标异常另查，不阻塞通知正文。
- [假成功消失后暴露真实错误] → 预期；便于排障。
- [type:3 无法深链] → 接受；深链靠可选 intent env。

## Migration Plan

1. 发布含修正的 push-service 镜像并 recreate。
2. 再触发 predict-imminent：期望通知栏可见；日志 `send_ok` 且若临时打 code 可见 `80000000`。
3. 若仍失败：看 `send_failed` 的 `code`，对照华为错误码与 token。
4. 回滚：回退 push-service 镜像。

## Open Questions

- （无阻塞）若产品后续要点击直达宝宝页，再定 `PUSH_HMS_CLICK_INTENT` 格式与客户端约定。
