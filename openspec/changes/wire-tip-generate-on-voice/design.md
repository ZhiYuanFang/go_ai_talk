## Context

`openspec/changes/tip-chat-streaming` 已实现：

- `api/v1/device_tip_http.go`：`POST /device/tip/generate`
- `internal/controller/device_tip.go`：`TipCtrl.Generate` + `writeSSEEvent`（`event:` + data 纯文本；done 携带 `{"answerId":"..."}`）
- voice 侧 `VoiceService.TipStream` → Python AI tip stream

但接线遗漏：

1. `register_voice_service.go` **未** `Bind(NewTipCtrl())`，TipCtrl 为孤儿控制器
2. `voice_route_proxy.go` 仅反代 `/voice/text/*`、`/voice/app/api/*`、`/voice/admin/api/*`，**无** `/device/tip/*`
3. Flutter `tip_repository.dart` URL 为 `$_baseUrl/tip/generate`（缺 `/device`），且按 `{type,content}` JSON 帧解析，与 Go SSE 方言不一致；models 使用错误的 `tipId int`

本变更承接 tip-chat-streaming **Decision 3（tip 宿主 = voice）**，只做接线与客户端对齐，不硬扩已勾完 tasks。

约束：AGENTS.md 服务边界、gateway-app 约定（路由+反代+Bearer+usage）、中文详细注释、不新增测试文件。

## Goals / Non-Goals

**Goals:**

- 将 TipCtrl 接到 voice-service 进程
- gateway 反代 `/device/tip/*` → VOICE_API_PROXY
- Flutter tip 路径与 SSE 方言对齐 Go；状态用 `answerId`
- usage：`POST /device/tip/generate` **计入统计**

**Non-Goals:**

- **不做** feedback 控制器 Bind / clinic tip feedback 完整飞轮（包 C `close-clinic-tip-feedback`）
- 不改 tip 宿主为 history（**锁定 voice**）
- 不改 SSE 方言为 Python JSON 帧（**锁定跟 Go**）
- 不新增 Redis、不改 DB、不新增测试文件
- 不修改已有 api/v1 tip 请求结构字段语义

## Decisions

### Decision 1: tip 宿主 = voice（锁定，不得改回 history）

**选择**：在 `RegisterVoiceServiceHTTP` 中 `group.Bind(NewTipCtrl())`（或等价），由 voice-service 直接服务 `POST /device/tip/generate`。

**理由**：

- 与 tip-chat-streaming Decision 3 一致：小贴士是 voice 专属（Python TipStream），与 history 域无关
- 减少 history→voice 多一跳；复用已实现的 `TipCtrl` / `TipStream`

**备选（否决）**：经 history 委派 —— 增加延迟与跨服务契约，且与既定 Decision 冲突，**本包禁止改回**。

### Decision 2: tip SSE 方言 = 跟 Go（锁定）

**选择**：服务端保持现有 `writeSSEEvent`：

```
event: thinking|answer|done|error
data: <纯文本或 done 的 JSON>
data: [DONE]
```

Flutter tip 解析对齐 chat SSE（`remote_feed_repository`）：读 `event:` 定类型，`data:` 为纯文本；`event: done` 时解析 `{"answerId":"..."}`。

**理由**：Go TipCtrl 已落地且与 clinic/chat 对外 SSE 风格一致；改 Flutter 成本低于回改 Go/Python。

**备选（否决）**：让 Go 改发 Python 风格 `{type,content}` JSON data —— 破坏既有 TipCtrl，本包不采纳。

### Decision 3: gateway 反代 `/device/tip/*` → voice

**选择**：在 `installVoiceProxyMiddleware`（或同类）增加 pattern `/device/tip/*`，复用 `VOICE_API_ROUTE_MODE` / `VOICE_API_PROXY_URL` / canary。

**理由**：App 路径已是 `/device/tip/generate`；gateway-app 必须放行到 voice，否则仅本地域服务注册仍对外不可达。

### Decision 4: Bearer 与 usage

**选择**：

- **不**将 `/device/tip/generate` 写入 `gateway_app_auth_exempt.go`（需登录）
- **不**写入 `usagestats/maintenance_skip.go`（**统计**，已拍板）

**理由**：与「新增 App 接口须确认 usage」结论一致；小贴士属用户主动能力，计入 usage。

### Decision 5: 本包不做 feedback（留给包 C）

**选择**：不 Bind clinic tip feedback 控制器；Flutter `submitFeedback` 可先把字段从 `tipId int` 对齐为 `answerId String`，但不要求 UI 飞轮与 Go feedback 路由完整可用。

**理由**：包边界清晰；避免与 `close-clinic-tip-feedback` 抢 scope。

### Decision 6: Flutter 状态标识用 answerId

**选择**：`TipContent` / `TipSSEEvent` 使用 `String? answerId`；流结束时从 done 事件写入；弃用错误的流内 `tipId int` 依赖。

**理由**：Go done 事件已是 `answerId` 字符串；int tipId 无法从当前流正确获得。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| gateway SSE 缓冲导致首包延迟 | 已有 `X-Accel-Buffering: no`；反代沿用现有 voice proxy 配置 |
| Flutter 改方言后与未部署 Go 不兼容 | 与 Go tip 上线同包发布；回滚可恢复旧仓库提交 |
| feedback 字段对齐后提交仍 404 | 预期：本包不 Bind feedback；完整闭环在包 C |

## Migration Plan

1. 部署 voice-service（含 TipCtrl Bind）+ gateway（含 `/device/tip/*` 反代）
2. 发布 Flutter：路径 + SSE 方言 + answerId
3. 回滚：gateway 去掉 tip pattern 或回退 voice Bind；客户端回退对应提交

## Open Questions

- 无（锁定决策已由包 B 输入拍板）
