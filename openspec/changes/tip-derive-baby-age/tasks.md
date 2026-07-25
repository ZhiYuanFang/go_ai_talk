## 0. 前置依赖



- [x] 0.1 确认 `d:\work\python_ai_talk\openspec\changes\fix-python-request-aliases` 已 apply（或同批次先合并），Go snake tip/intent/clinic body 不再因 alias 422



## 1. Python：派生月龄与当前时间



- [x] 1.1 在 `d:\work\python_ai_talk\app\tip\schemas\tip.py` 删除（或废弃且非必填并忽略）`TipRequest.baby_age_months`、`TipRequest.current_time`；更新中文注释

- [x] 1.2 调整 `app/api/routes/tip.py`：initial_state 不再从 request 读取上述两字段；注释去掉「请求必含月龄/时间」

- [x] 1.3 在 `fetch_baby_profile` 之后（tip 图节点或同文件扩展）按 design 决策 1 用 `Asia/Shanghai` 自算月龄；无生日 → 内部 `None` + 提示词「未知」（不得用 0 表示未知）

- [x] 1.4 修改 `app/tip/graphs/states/tip_state.py` / 相关节点：`baby_age_months` 可为可选；`current_time` 改为写 prompt 时生成或不再作为请求驱动状态

- [x] 1.5 修改 `app/tip/graphs/nodes/prompts/tip_answer.py` 与 `stream_tip_response.py`：提示词时间用 `Asia/Shanghai` 可读时间（MAY 附 Unix 秒）；月龄文案区分「n 个月」与「未知」

- [x] 1.6 若 tip 知识检索使用月龄：无月龄时跳过月龄过滤 / 宽检索（不得用 0 冒充）



## 2. Go：Tip API / TipStream 瘦身



- [x] 2.1 修改 `api/v1/device_tip_http.go`：删除 `BabyAgeMonths`、`CurrentTime` 字段及假注释

- [x] 2.2 修改 `internal/controller/device_tip.go`：调用 `TipStream` 时不再传月龄/时间

- [x] 2.3 修改 `internal/services/contracts/runtime_contracts.go` 与 `internal/services/voice/voice_chat.go`：`TipStream` 签名去掉 `babyAgeMonths`、`currentTime`；删除 `currentTime<=0` 补 `time.Now()` 逻辑

- [x] 2.4 修改 `internal/services/voice/python_ai_client.go`：`TipStreamRequest` 删除 `BabyAgeMonths`、`CurrentTime` 及 json 标签

- [x] 2.5 全文搜索构造 `TipStream` / `TipStreamRequest` / `DeviceTipGenerateReq` 的调用点并改干净；同步修正 `openspec/changes/tip-chat-streaming` 等文档中仍列 `babyAgeMonths`/`currentTime` 为必填的表述



## 3. Flutter：去掉传参



- [x] 3.1 修改 `d:\work\flutter_ai_talk\app\lib\data\tip_repository.dart`：`streamTip` 去掉 `babyAgeMonths`；body 去掉 `babyAgeMonths`、`currentTime`

- [x] 3.2 修改 `app/lib/providers/tip_provider.dart`：`startStreaming` 去掉 `babyAgeMonths`

- [x] 3.3 修改 `app/lib/ui/home_screen.dart`：`_triggerTipGeneration` 删除月龄计算与传参



## 4. 联调与文档



- [x] 4.1 端到端：有生日设备 → 提示词含正确月龄；无生日 → 「未知」且不出现「0 个月」伪装

- [x] 4.2 确认 Chat 宿主未改动；Go `PythonAIClient` JSON 仍为 snake_case

- [x] 4.3 本阶段不新增 `*_test.go` / `*_test.dart` 测试文件

