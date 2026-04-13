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

### 4) 页面能力

- 设备管理页：`/device/admin`
- 设备历史页：`/device/history/{deviceNo}`
  - 展示事件历史
  - 独立模块展示历史成长建议

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

---

### B. 设备管理 API

#### 1. 管理页

- **GET** `/device/admin`

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

---

### C. 设备历史与建议 API

#### 1. 历史页面

- **GET** `/device/history/{deviceNo}`

#### 2. 获取设备历史事件

- **Method**: `GET`
- **Path**: `/device/history/api/list`
- **Query**: `deviceNo`

示例：

`/device/history/api/list?deviceNo=device-001`

#### 3. 获取设备成长建议历史

- **Method**: `GET`
- **Path**: `/device/history/api/suggest`
- **Query**: `deviceNo`

示例：

`/device/history/api/suggest?deviceNo=device-001`

---

## 三、主要配置

运行配置见 [manifest/config/config.yaml](manifest/config/config.yaml)：

- `server.address`：服务监听地址（默认示例 `:9701`）
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

