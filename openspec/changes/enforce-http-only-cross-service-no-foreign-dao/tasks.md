## 1. 契约与 device HTTP 面

- [x] 1.1 盘点 `voice` 对 `DeviceAdmin` / `DeviceProfile` 的全部调用点，列出所需 device HTTP 路径（读/写事件、动作、画像、注册校验等）。
- [x] 1.2 在 `internal/services/contracts` 与 device controller 中补齐缺失的 **internal** 路由（与 admin 口令隔离），并与 `http_targets` 对齐。
- [x] 1.3 定义 voice 侧 **Device 域访问接口**（可拆分多个小接口），与 `DeviceProfileContract` 的 local/remote 策略一致。

## 2. Voice 侧适配

- [x] 2.1 实现 remote HTTP 客户端（错误壳解析、超时、与现有 `responseEnvelope` 一致）。
- [x] 2.2 将 `domain_refs`（或等价入口）改为返回上述接口，默认生产走 remote；local 仅指向 device 基址的 HTTP。
- [x] 2.3 替换 `voice_chat_understanding` 等对 `DeviceAdmin()` 的进程内调用为契约调用。

## 3. 配置与部署

- [x] 3.1 更新 `config.voice-service.yaml` 与 `manifest/deploy/.../voice-deployment.yaml`：`DEVICE_SERVICE_URL` 及模式类环境变量必填说明。
- [x] 3.2 更新 `docs/runbooks/release-deploy-and-run.md`：单服务单库 + 仅 HTTP 跨域的检查项。

## 4. 治理

- [x] 4.1 更新 `AGENTS.md`：明确 voice 包禁止他域 `dao`；评审 grep 要点。
- [x] 4.2 手工验证：仅 voice 库配置下语音主路径不触 device 表直连（可结合日志/SQL 审计）。
