## Why

`add-pangbao-ai-clinic-room` 与 `amend-pangbao-clinic-ux` 已交付胖宝诊疗 WS 流式问答与 session 恢复，但用户无法在 DeepSeek **思考/回答流式过程中**停止、改问或离开页面时显式取消；当前 WS 读循环同步阻塞 `HandleQuestion`，Flutter 在 `_busy` 时锁输入，页面退出仅被动关闭连接，后端依赖写失败才中断 LLM，导致资源浪费、额度/限流语义不清、UX 与竞品「可打断 AI」预期不符。

## What Changes

- **新增** 客户端生成 **`turnId`（UUID）**  per 提问；上行 `question` 含 `{text, turnId}`，上行 **`cancel`** 含 `{turnId}`。
- **新增** 下行 **`turn_cancelled`** 帧 `{turnId, reason: superseded|cancelled|disconnected}`；所有流式下行帧（`thinking_delta` / `answer_delta` / `answer_done`）**MUST** 携带 `turnId`。
- **修改** voice-service：`HandleQuestion` 异步 goroutine；连接级 `activeTurn` + `context.CancelFunc`；读循环非阻塞；收到 `cancel`、新 `question` supersede、WS 关闭时取消 LLM ctx 并下发 `turn_cancelled`。
- **修改** 额度与限流语义：**cancelled / superseded** 轮次 **MUST NOT** consume `clinic_ai`；限流计数 **仅** 在 **`answer_done` 成功** 后计入（supersede 不额外扣费）。
- **修改** session：仅 **`answer_done` 完成** 的 Q&A 写入 Redis；被取消/被 supersede 的 partial turn **MUST NOT** append。
- **修改** Flutter（`flutter_ai_talk`）：流式期间输入可用；发送新问或点停止 **supersedes** 当前 turn；`dispose` / App 后台 **先** `cancel` 再 tearDown WS；按 `turnId` 过滤 stale delta；可选 tap 用户气泡编辑改问。
- **不新增** 测试文件、背景 ticker、Redis 键、网关路径变更。

## Capabilities

### New Capabilities

（无——本变更为既有胖宝诊疗能力的协议与 UX 增量。）

### Modified Capabilities

- `pangbao-ai-clinic`：`turnId` 协议、`cancel` / `turn_cancelled` 帧、异步处理与显式取消、额度/限流/session 对未完成轮次的语义。
- `pangbao-ai-clinic-flutter`：流式中断/改问 UI、`cancel` 生命周期、stale 帧过滤、页面退出显式取消。

## Impact

- **go_ai_talk**
  - `internal/controller/voice_clinic_ws.go`：读循环 dispatch、`cancel` 帧、连接关闭 cancel active turn
  - `internal/services/voice/clinic_*.go`：`HandleQuestion` 可取消 ctx、下行帧带 `turnId`、`turn_cancelled`、quota/rate/session 边界
  - OpenSpec 变更增量（本 change）
- **flutter_ai_talk**（`d:\work\flutter_ai_talk`）
  - `clinic_ws_client.dart`：`turnId`、`cancel` 发送、解析 `turn_cancelled` 与带 `turnId` 的 delta
  - `pangbao_ai_screen.dart`：流式期间可输入/停止/改问、`dispose`/后台 cancel、stale 过滤
- **依赖**：建立在 `add-pangbao-ai-clinic-room`、`amend-pangbao-clinic-ux`（`session_sync`、thinking UI）之上；跨仓 go_ai_talk + flutter_ai_talk
- **App API usage 统计**：无新 HTTP 接口，无需 maintenance_skip 变更
