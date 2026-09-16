## 1. loggercfg 平台包

- [x] 1.1 新增 `internal/platform/loggercfg`：`ApplyFromEnv(service string)`，读 `GF_LOGGER_LEVEL`；空则 return；先触达默认 logger 再 `SetLevelStr`；成功 Warning、失败 Error（中文注释）
- [x] 1.2 在下列服务启动路径（`prepare*Runtime` 末尾或等价位置，须在可能初始化 logger 的 dbcfg/rediscfg 之后）调用 `loggercfg.ApplyFromEnv`：`push-service`、`voice-service`、`device-service`、`history-service`、`ucg-service`、`cash-service`、`gateway-app-server`、`notify-service`、`sim-user-service`、`mcp-service`；若存在独立 gateway `main` 一并接入

## 2. push 可观测 Warning

- [x] 2.1 `internal/controller/push/internal_push.go`：鉴权失败 / 校验失败 / 受理成功打 Warning（`[push]`，含 wxId/bizType/badge/silent/alertLen；无 token、无 alert 全文）
- [x] 2.2 `push_async.go`：recover 打 Error/Warning，禁止空吞 panic
- [x] 2.3 `push_dispatcher.go`：dispatch 时打 Warning（wxId/bizType/deviceCount）；`send_ok` Warning；保留既有 skip/send_failed Warning
- [x] 2.4 APNs/HMS/MiPush：凭证未配置由 Debug 改为 Warning

## 3. 运维注释与 compose 说明

- [x] 3.1 更新 `manifest/docker/.env.example` 中 `GF_LOGGER_LEVEL` 注释：留空 / prod / info / all（及非法值）效果、需 recreate 容器
- [x] 3.2 更新 `manifest/docker/env/.env.prod`（及已含该键的 `.env.test` / `.env.release` / `.env.local`）第 40 行附近同语义注释，便于随时改值
- [x] 3.3 修正 `docker-compose.microservices.yml` 中 `x-logger-env` 注释：经 `loggercfg.ApplyFromEnv` 生效，非 GoFrame 自动覆盖 yaml

## 4. 自检

- [x] 4.1 本地或容器：设 `GF_LOGGER_LEVEL=prod` 启动 push/voice，确认有 `[loggercfg] ... applied`，且普通 Infof 不再出现
- [x] 4.2 触发一次 internal by-biz-type（或 predict-imminent）：确认 push 侧可见 `accepted` / `dispatch` 或 `no_device` / credentials / `send_ok` 等 Warning
- [x] 4.3 确认未新增 `*_test.go`；`hack/check-service-import` 若适用仍通过
