# llm-care-alert-daily — Go 契约（来自 Flutter 主变更）

主规格与任务在兄弟仓 `flutter_ai_talk`：`openspec/changes/llm-care-alert-daily/`。

本文件供 Go 实现对照；**handler 已落地于 voice-service**。

## 职责

- 鉴权后按宝宝日缓存护理留意列表；编排调用 Python KG+LLM。
- 触发者账号经 VIP∪`care_alert` 额度判定 premium → `careAlert` lane 正式模；非 premium → lane free（可空则 omit 由 Python 自选）。VIP 查询失败降级为非 VIP（仍可有额度）。详见变更 `vip-quota-joint-entitlement`。
- **不**扣 clinic 配额；**不**支持纯设备会话（必须 `wxId > 0`）。
- 忽略：删当日缓存项；飞轮固定意图（无 NLP）。

## HTTP（经 gateway `/device/api/...`）

宿主：`voice-service`（与 tip/clinic 同进程）。  
网关：`installVoiceProxyMiddleware` 绑定 `/device/api/care-alert/*`。  
鉴权：gateway-app Bearer（**非**白名单，须登录）；网关注入 `X-Internal-Wx-Id`（`wxId > 0` 必填）；`deviceNo` 由客户端 query/body 传入（宝宝维度）。  
VIP 真相源见变更 `vip-cash-service`（`cash-service` / `ai_voice_cash`）；已 supersede `care-alert-vip-by-wx` 的 `wx.is_vip` 路径。

### GET `/device/api/care-alert/daily`

- Query：`deviceNo`；Header：`X-Internal-Wx-Id`（必填，>0）
- 缓存键：`cachekit.CareAlertDailyKey` → `voice:carealert:daily:{deviceNo}:{YYYY-MM-DD}`（`Asia/Shanghai`；builder 会规范化 segment）
- 命中：返回相同 `items`，**不**重跑 LLM
- 未命中：进程内 single-flight + Redis lock（`voice:carealert:lock:{deviceNo}:{day}`）阻塞生成后写入缓存并返回
- 成功 `data`：

```json
{
  "day": "2026-08-08",
  "items": [
    {
      "suggestionId": "<uuid>",
      "eventId": "...",
      "eventName": "...",
      "summaryLine": "...",
      "followUpPrompt": "...",
      "reasons": [
        {
          "type": "elongatedInterval|longActive|suddenAbsence|<ext>",
          "score": 1.0,
          "expectationUsed": true,
          "ageMonths": 3,
          "medianGapMs": 0,
          "lastGapMs": 0,
          "expectGapMaxMs": 0,
          "p75DurMs": 0,
          "elapsedMs": 0,
          "expectDurMaxMs": 0,
          "dailyAvg": 0.0,
          "recent48hCount": 0,
          "stillExpected": true,
          "detailLines": []
        }
      ]
    }
  ]
}
```

### DELETE `/device/api/care-alert/daily/item`

- Query（Flutter 客户端）：`deviceNo`, `suggestionId`；Header：`X-Internal-Wx-Id`（必填，>0）
- 仅删除**当日**缓存中该 `suggestionId`；次日可再出现
- 成功返回更新后 `{day, items}`（HTTP 200 + envelope `code=0`；非 204，以兼容 Flutter `deleteEnvelope`）

### POST `/device/api/care-alert/feedback`

```json
{
  "deviceNo": "...",
  "suggestionId": "<uuid>",
  "intent": "ignore|follow_up"
}
```

Header：`X-Internal-Wx-Id`（必填，>0）。固定意图飞轮；**不得**对自由文本做 NLP。本地落日志 + 尽力转发 Python `/v1/care-alert/feedback`；Python 失败时不阻断客户端。

## Go → Python

与 tip/clinic 同约定，挂在 Python `APIRouter(prefix="/v1")` 下（**不是** `/internal/...`）：

- 分析：`POST {pythonAiTalk.url}/v1/care-alert/analyze`
- 飞轮：`POST {pythonAiTalk.url}/v1/care-alert/feedback`（ACK；Go **best-effort**，失败仅打 Warning）
- 请求体字段见 `PythonAIClient.CareAlertAnalyze` / `CareAlertFeedback`（snake_case + `model` 简写 + `model_cfg`）

## 实现提示

- Redis 缓存 MUST 经 `cachekit`；键经 builder（`CareAlertDailyKey` / `CareAlertDailyLockKey`）。
- 每项生成 UUID `suggestionId`；缺 `followUpPrompt` 时 Go 用 summary/事件名补齐。
- 权益：`ResolveLaneModel(..., LaneCareAlert, care_alert, Account)`；VIP 经 cash `GET /cash/internal/api/vip/by-wx-id`；独立 feature `care_alert`；成功时**仅非 VIP** consume。**禁止**按 `deviceNo` 判 VIP，**禁止**纯设备会话，**禁止** DeepSeek/Zhipu 硬切。
- 选模/并发：`careAlert` lane（Admin 可配正式模 + free）；日缓存仍按 `deviceNo+日`（触发者权益：仅 miss 生成读触发者权益）。

## 状态

- [x] GET daily + 缓存
- [x] DELETE item
- [x] POST feedback
- [x] VIP∪care_alert 额度选模 + 调 Python；非 VIP 成功扣 care_alert；不扣 clinic 配额

## Flutter 备注

主变更任务 **6.2**（手工路径：加载隐藏 → 跑马灯 → 忽略/追问）仍为手动验收，未自动勾选。
