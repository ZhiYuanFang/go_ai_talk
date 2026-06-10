# GoFirst 智能语音服务

基于 GoFrame 的语音对话服务，核心包含：

- 语音 WebSocket 对话（ASR + LLM + TTS）
- 设备管理后台（注册设备、查看设备列表）
- 设备历史记录页面（事件历史 + 成长建议）
- 事件记录能力（开始/结束、数量、备注）
- 每设备每日成长建议缓存（同日重复问询直接返回数据库结果）

---

## 一、功能概览

### 1) 对话能力

- **语音对话**：通过 WebSocket 上行 PCM 音频分片，服务端返回合成语音分片。
- **多轮上下文**：按 `deviceNo` 维度缓存对话上下文（进程内缓存，服务重启后丢失）。

### 2) 事件记录能力

对话中会尝试识别与 `event` 表匹配的事件意图，自动记录到 `history` 表：

- 识别动作：`start` / `end`
- `end` 且无已开始记录时：自动补一条“开始=结束时间”的记录
- 结束动作可提取数量写入 `event_number`
- 若缺少数量且事件需要数量：会追问，下一轮补录
- 备注信息写入 `remark`

### 3) 成长建议能力

- 用户询问成长建议时，系统会结合设备历史 + 生日信息生成建议
- 每设备每天仅生成一次并落库 `suggest`
- 同一天再次询问直接返回当日已生成建议

### 4) Web 页面与静态入口（汇总）

仓库内联调/运维用页面集中在 `resource/public/`，由 **App 专用网关**（`gateway-app-server`，默认 `:9702`）提供；主网关（`:9701`）对 admin/history 壳页 **302** 至 `GATEWAY_APP_PUBLIC_BASE_URL` 对应路径。下文路径均相对于 App 网关根地址（如 `http://127.0.0.1:9702`）。

| 进程 | 路径 | 静态文件 | 用途简述 |
|------|------|----------|----------|
| 主网关 | `/` | （内联文案） | 根路径提示「智能语音」；访问 `/device/admin*` 302 至 App 网关 |
| App 网关 | `/device/admin` | `resource/public/admin.html` | **运维 Hub**：`GATEWAY_APP_ADMIN_*` 登录后签发 Admin JWT，各模块共用 |
| App 网关 | `/device/admin/qa-records` 等 | 见 `admin-modules.js` | 问答库、反馈、API 统计、UCG 管理等子页（须 Hub 登录） |
| App 网关 | `/device/history/{deviceNo}` | `resource/public/history.html` | 单设备历史事件 + 成长建议；壳页内请求 `/device/history/api/*` |
| App 网关 | `/device/app/integration-test.html` | `resource/public/gateway-app-integration-test.html` | App 联调（Bearer 白名单页，**勿暴露公网**） |
| App 网关 | `/device/app/version-admin.html` | `resource/public/gateway-app-version-admin.html` | 版本与 APK 管理（Admin JWT，与 Hub 同一登录） |
| App 网关 | `/device/app/apk/{文件名}` | （磁盘 `/apk/ai_talk/` 等配置目录） | 安装包匿名下载；URL 由版本管理上传流程写入库表 |
| 两网关 | `/resource/*` | `resource/` 下文件 | GoFrame 静态资源目录（按需访问） |
| App 网关 | `/` | （内联文案） | 根路径提示「智能语音 App 网关」 |

说明：

- **Web 管理唯一入口**为 App 网关 `:9702`（测试 Nginx 反代或 `:19702`）；主网关不再挂载 admin 静态页。
- Hub 登录 `POST /device/admin/api/login` 返回 Admin JWT；管理 API 请求须 `Authorization: Bearer`；网关校验后向下游注入 `X-Admin-Password`（`DEVICE_ADMIN_PASSWORD` / `UCG_ADMIN_PASSWORD`）。

---

## 二、API 总览

### A. 对话 API

#### 1. 语音对话（WebSocket）

- **Path**: `/voice/chat/ws`
- 协议顺序：`start` -> `binary audio chunk` -> `end`

客户端首条文本（必须）：

```json
{"type":"start","deviceNo":"device-001","sampleRate":16000,"bits":16,"channels":1,"length":32000}
```

客户端结束文本：

```json
{"type":"end"}
```

服务端返回：

```json
{"type":"audio_chunk","audio":"<base64片段>","sampleRate":16000}
```

```json
{"type":"audio_end","code":0,"exit":false}
```

若识别到退出意图：

```json
{"type":"exit","code":0,"exit":true}
```

错误返回（Text JSON）：

```json
{"type":"error","code":1,"stage":"chat","message":"..."}
```

#### 2. 实时听写（WebSocket，仅语音→中文文本）

- **Path**: `/voice/asr/ws`（经主网关 `:9701` 或 App 网关 `:9702` 透传至 voice-service，与对话 WS 共用 `VOICE_WS_ROUTE_MODE` / `VOICE_WS_PROXY_URL`）
- **用途**：仅需流式听写（`asr_partial` / `asr_final`），无 LLM、TTS、对话落库
- **协议**：`start` → 二进制 PCM 分片 → **`commit`（必须，一句结束）** → `end`（关会话）
- **截句**：服务端**不会**因静音自动结束一句；`asr_final` **仅**在 `commit` 或 `end` 时下发（`source` 为 `client` 或 `end`）
- **预览纠正**：说话过程中百度引擎可能对某短语产生 final，服务端会**再发一条** `asr_partial`（与中间 partial 同类型）覆盖预览，前端用最新 `text` 替换即可；**定稿**仍须 `commit`

客户端 `start`（`mode` 可省略或设为 `stream`）：

```json
{"type":"start","deviceNo":"device-001","sampleRate":16000,"bits":16,"channels":1,"length":32000}
```

松手或点「完成」时**必须**发送（否则只有预览、没有定稿听写）：

```json
{"type":"commit"}
```

结束听写会话：

```json
{"type":"end"}
```

服务端下行示例：

```json
{"type":"started","code":0,"mode":"stream"}
```

```json
{"type":"asr_partial","code":0,"text":"你好"}
```

```json
{"type":"asr_final","code":0,"text":"你好世界","source":"client"}
```

---

### B. 设备管理 API

#### 1. 管理页（Web）

- **GET** `/device/admin` → `resource/public/admin.html`（**仅 App 网关** `:9702`；主网关 `:9701` 302 至 App 网关）

#### 2. 注册设备

- **Method**: `POST`
- **Path**: `/device/admin/api/register`
- **Header**: `X-Admin-Password: a521521521`
- **Body**:

```json
{"deviceNo":"device-001"}
```

#### 3. 设备列表

- **Method**: `GET`
- **Path**: `/device/admin/api/list`
- **Header**: `X-Admin-Password: a521521521`

#### 4. 事件管理（logo / color）

- **列表**: `GET /device/admin/api/event/list` — 每项含 `logo`（CDN 绝对 URL，如 `https://resorce.cuplay.top/event/...`）、`color`（如 `#FF6B6B`）
- **新增/更新**: `POST` multipart 至 `/device/admin/api/event/add` 或 `/event/update`
  - 表单字段：`name`、`eventType`（`number` 计数 / `time` 计时 / `one` 一次性）、`extraNames`、`color`（`#RGB`/`#RRGGBB`）、可选文件 `logo`
  - 更新时未传 `logo` 文件则保留原 logo
- **存储**：库内 `event.logo` 为 OSS objectKey（`event/...`）；API 响应返回 CDN URL。上传经 ucg-service OSS；详见 `docs/runbooks/event-logo-oss-migration.md`。
- **管理页**：`admin.html` 事件表展示 Logo、色调；**点击**色块或 Logo 可直接更新
- **BREAKING**：`GET /device/app/api/version/check` 的 `downloadUrl` 仍为**路径**（非绝对 URL），App 须自行拼接网关基址。

---

### C. 设备历史与建议 API

#### 1. 历史页面（Web）

- **GET** `/device/history/{deviceNo}` → `resource/public/history.html`（主网关或 App 网关）

#### 2. 获取设备历史事件

- **Method**: `GET`
- **Path**: `/device/history/api/list`
- **Query**: `deviceNo`, `page`, `pageSize`
- **Response**: `list`, `total`, `page`, `pageSize`

示例：

`/device/history/api/list?deviceNo=device-001&page=1&pageSize=20`

#### 3. 获取设备成长建议历史

- **Method**: `GET`
- **Path**: `/device/history/api/suggest`
- **Query**: `deviceNo`

示例：

`/device/history/api/suggest?deviceNo=device-001`

---

## 三、主要配置

运行配置见 [manifest/config/config.yaml](manifest/config/config.yaml)（主网关）；**App 网关**独立配置见 [manifest/config/config.gateway-app-server.yaml](manifest/config/config.gateway-app-server.yaml)（含 `database.app`、JWT、版本管理页相关 `gatewayApp.publicBaseUrl` 等）。微服务下 `voice-service` / `history-service` 的对话参数统一在 [manifest/config/voice-chat.shared.yaml](manifest/config/voice-chat.shared.yaml)：

- `server.address`：服务监听地址（主网关默认示例 `:9701`，App 网关默认 `:9702`）
- `voiceChat.audio`：音频参数与最大时长
- `voiceChat.session`：会话轮数、TTL、清理间隔
- `voiceChat.stt`：语音识别配置（百度）
- `voiceChat.deepseek`：大模型配置
- `voiceChat.tts`：语音合成配置（百度）

---

## 四、常用命令

- `make cli`：安装/更新 GoFrame CLI
- `make dao`：按 `hack/config.yaml` 生成 DAO/DO/Entity
- `make service`：生成 service 接口文件
- `go fmt ./... && go build ./...`：格式化并编译检查

---

## 五、说明

- 对话上下文缓存为进程内内存态，重启后丢失是预期行为。
- 建议在生产环境通过环境变量/密钥管理替代硬编码管理口令与三方密钥。

