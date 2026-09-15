## 1. 基础设施与键登记

- [x] 1.1 RabbitMQ：启用 `rabbitmq_delayed_message_exchange`（镜像/插件文档）；更新 `hack/rabbitmq-init`（及 ps1）声明延时交换机（如 `voice.delayed`）、队列（如 `voice.predict.imminent.q`）与绑定
- [x] 1.2 在 `internal/platform/cachekit/keys_*.go` 登记待办列表键、推送去重键（中文注释：TTL/失效/跨进程语义）；经 `cachekit.Default()` 访问
- [x] 1.3 `eventkit`：支持向 delayed exchange 发布带 `x-delay` 的消息（或 voice 侧专用 publisher），routing key 与 design 一致

## 2. device：按 device_no 列全部 wxId

- [x] 2.1 device-service 实现 `ListWxIdsByDeviceNo`（含上限/截断告警）；api + controller + 服务逻辑
- [x] 2.2 `clients/device` 与 contracts 路径补齐，供 voice 调用

## 3. 公共推送契约

- [x] 3.1 抽出/新增 internal 推送 API：`wxId` + `bizType` + 文案/data；`predict_imminent` 不走 UCG 社区角标语义
- [x] 3.2 `clients/ucg`（或约定 clients 包）封装调用；禁止 voice import `services/ucg`
- [x] 3.3 确认 token 注册复用路径；文档/注释说明客户端需注册才能离线收到

## 4. voice：同步 API 与 Redis/延时投递

- [x] 4.1 新增 api/v* 请求响应与 controller（登录+绑机校验；最后写入赢全量替换）
- [x] 4.2 实现 Redis 待办读写与同 device 并发安全（短锁或等价）
- [x] 4.3 同步成功后按条计算 `x-delay` 并 Publish；空列表只清 Redis 不发取消消息
- [x] 4.4 gateway-app：反代新路径；向负责人确认 usage 是否计入后再处理 `maintenance_skip` / 白名单（**代理已接；usage 未改 skip，待负责人确认**）

## 5. voice：延时消费与闸门

- [x] 5.1 `voice-service` 启动预测临近 AMQP consumer（配置开关）；handler 业务路径一律 Ack
- [x] 5.2 消费：Redis 匹配 → 5 分钟去重 → history 30 分钟（根展开叶子）→ `ListWxIdsByDeviceNo` → 公共 push
- [x] 5.3 推送失败只打日志并 Ack；禁止 Nack requeue 与消费内二次延时 Publish
- [x] 5.4 中文业务注释覆盖同步、延时、消费校验与失败语义

## 6. 联调与文档

- [x] 6.1 本地/测试环境验证：替换列表后旧延时消息不推；过期仍推；已有 history 不推；双 wx 扇出（**本会话以 `go build` 相关包通过为准；端到端需启 delayed 镜像 + init 后手工验收**）
- [x] 6.2 更新部署/runbook：插件启用、队列名、consumer 开关、失败不重试说明
- [x] 6.3 （孪生仓）Flutter：预测更新后调同步 API；确认推送注册与深链（本仓见 `CLIENT.md`）
