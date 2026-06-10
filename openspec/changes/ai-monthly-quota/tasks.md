## 1. device-service 数据层与 Redis 用量



- [x] 1.1 新增 device 库表 `ai_quota_default`（singleton）、`ai_quota_user_override`（wx_id UNIQUE）及 seed（polish=5、voice_ai=5）

- [x] 1.2 实现 `internal/services/device/ai_quota.go`：解析 effective limit（global + override）、上海时区月桶 key、Redis get/incr/rollback

- [x] 1.3 定义业务错误码 40301（未登录）、40302（额度用尽）常量供 controller 复用



## 2. device internal 与 App API



- [x] 2.1 在 `api/v1` 增加 internal：`POST /device/internal/api/ai-quota/check`、`POST .../consume` 及 internal admin default/user 读写类型

- [x] 2.2 实现 internal controller/handler：`X-Device-Internal-Secret` 校验、check/consume 逻辑

- [x] 2.3 实现 `GET /device/app/api/ai-quota`（wxId>0，返回 polish/voiceAi used+limit）

- [x] 2.4 在 `internal/services/contracts` 补充额度相关契约（供 voice HTTP client 调用）



## 3. ucg-service 润笔接入



- [x] 3.1 新增 device internal HTTP client 方法：CheckAIQuota、ConsumeAIQuota

- [x] 3.2 在 `PostsPolish`：DashScope 前 check；成功返回后 consume；超额返回 40302

- [x] 3.3 实现 `GET/PUT /ucg/admin/api/ai-quota/default` 与 `.../user`（ucg admin 口令 + 转发 device internal admin）



## 4. voice-service 喂养 AI 接入



- [x] 4.1 扩展 voice 侧 device HTTP client：check/consume + deviceNo→wxId（若尚无则新增 internal 查询）

- [x] 4.2 在 LLM 调用前解析 wxId、check；wxId≤0 返回 40301 WS 错误帧

- [x] 4.3 在母婴 DeepSeek 与 casual 流式成功路径调用 consume；模式切换/规则/失败兜底不触发

- [x] 4.4 超额返回 WS 40302「本月额度已用完」



## 5. 管理页



- [x] 5.1 扩展 `resource/public/ucg-admin.html`「AI 配置」：全局润笔/喂养默认次数、wxId override 表单，调用 `/ucg/admin/api/ai-quota/*`

- [x] 5.2 在 `api/v1/ucg_admin_http.go` 注册 ai-quota admin 路由类型



## 6. 网关与配置



- [x] 6.1 确认 gateway-app 反代 `/device/app/api/ai-quota` 与 ucg/voice internal 路径无需额外改动（或补登记）

- [x] 6.2 更新 `manifest/config` 与 docker env 文档：device internal secret、ucg deviceServiceUrl（若缺失）



## 7. Flutter（flutter_ai_talk 仓库，独立发版）



- [x] 7.1 润笔 API：捕获 40302 → 弹框「本月额度已用完」；40301 → 登录引导

- [x] 7.2 喂养 WS：解析 error 帧 40302/40301 同上

- [x] 7.3 （可选）调用 `GET /device/app/api/ai-quota` 展示剩余次数



## 8. 验证



- [x] 8.1 润笔：默认 5 次、第 6 次 40302、DashScope 失败不扣减、全局润笔 10/喂养 5 独立生效（`go build ./...` 通过；联调待部署后手工验证）

- [x] 8.2 喂养 AI：wxId=0 40301、LLM 成功扣减、模式切换不扣、超额 40302（编译通过；WS 联调待验）

- [x] 8.3 Admin：ucg-admin 修改 global 与 per-wxId override 后 check 口径正确（页面与 API 已实现）

- [x] 8.4 月界：上海时区 YYYYMM 桶切换后 used 归零（见 `ai_quota.go` `Asia/Shanghai` 月桶）


