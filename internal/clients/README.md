# clients 排除项说明（service-package-isolation）

本目录 `internal/clients/{cash,device,history,ucg,voice}` 仅承载**本仓库域服务**之间的出站 HTTP 客户端。

## 不迁入域名 clients 的依赖

| 类别 | 位置（示例） | 原因 |
|------|--------------|------|
| Python AI HTTP | `internal/services/voice/python_ai_client.go` | 外部推理进程，非 Go 域服务 |
| 阿里云内容安全 Green | `internal/services/ucg/green_client.go` | 云厂商 SDK/API |
| aimodel 选模 | `internal/services/aimodel` | 共享基础设施，设计允许跨进程 import |
| 设备域本地 DAO 适配 | `services/device` 内 local/canary | 同进程实现，非出站客户端 |

若新增「调他域 Go 服务」的 HTTP 客户端，应落在 `internal/clients/{被调服务}`，禁止挂在被调方 `services/{X}` 再被调用方 import。
