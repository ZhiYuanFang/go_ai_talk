## Why

小贴士链路当前要求 Flutter/Go 传入 `babyAgeMonths` / `currentTime`（及 Go→Python 的 `baby_age_months` / `current_time`），但：① 月龄本可由 Python 在拉取宝宝画像后按生日计算，客户端传值易漂移且与「0 则服务端推算」假注释不符（Go 实际未按生日推算）；② 当前时间应由 Python 写提示词时生成，无需跨端传递。在 `fix-python-request-aliases` 修好 snake 入站后，应立即瘦身 tip 契约，消除冗余字段与错误语义。

## What Changes

- **依赖**：本变更依赖 `python_ai_talk` 的 change `fix-python-request-aliases`（Go snake body 可过 Python 校验）完成后再实施。
- **Python**：`fetch_baby_profile` 之后根据 birthday **自算月龄**写入 tip 状态/提示词；写提示词时用本地时间（`time.time()` 或 Asia/Shanghai 可读时间）生成 **current now**；`TipRequest` **删除/废弃**请求字段 `baby_age_months`、`current_time`（不再必填）。
- **Go**：**BREAKING**（接线期 tip 契约）`DeviceTipGenerateReq`、`TipStream`、`TipStreamRequest` 去掉月龄与 currentTime 相关字段/参数；删除「0 则服务端推算」等假注释与伪兜底逻辑。
- **Flutter**：`tip_repository` / `tip_provider` / `home_screen` 去掉传月龄与 `currentTime`。
- **不改** Chat 宿主（仍走 history SSE）；App 其它接口 camel 约定不变。

## Capabilities

### New Capabilities

- `tip-python-derived-context`：Python tip 图在画像拉取后自算月龄，并在写提示词时生成当前时间上下文；无生日时语义为「未知」而非 0 个月。
- `tip-contract-slim`：跨仓 tip 请求契约瘦身——Go App API / voice TipStream / Python TipRequest / Flutter tip 调用去掉 `babyAgeMonths`/`currentTime`（及 snake 对等字段）。

### Modified Capabilities

- （无）`openspec/specs/` 下无已归档 tip 能力主规格；既有 change 文档（如 `tip-chat-streaming`）中的请求字段描述由本 change tasks 同步修正，不作为 delta 主规格。

## Impact

- **Python**（`d:\work\python_ai_talk`）：`app/tip/schemas/tip.py`、`app/api/routes/tip.py`、`tip_state` / tip graph nodes / `tip_answer` 提示词、`fetch_baby_profile` 后衍生逻辑。
- **Go**（本仓）：`api/v1/device_tip_http.go`、`internal/controller/device_tip.go`、`internal/services/voice/voice_chat.go`（`TipStream`）、`python_ai_client.go`（`TipStreamRequest`）、`contracts` 中 `TipStream` 签名；相关 openspec 假注释/字段列表。
- **Flutter**（`d:\work\flutter_ai_talk`）：`app/lib/data/tip_repository.dart`、`app/lib/providers/tip_provider.dart`、`app/lib/ui/home_screen.dart`。
- **非影响**：Chat / history SSE 宿主；clinic 请求体（除可对齐的时区约定外）。
