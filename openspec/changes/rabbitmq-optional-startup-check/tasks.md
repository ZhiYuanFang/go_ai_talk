## 1. runtimecheck

- [x] 1.1 `dependencies.go`：增加 `DependencyOptions.RequireRabbitMQ`；false 时 MQ 探活 warn-only

## 2. 服务入口

- [x] 2.1 gateway（`internal/cmd/cmd.go`）、gateway-app、device、history、voice、ucg：`RequireRabbitMQ: false`
- [x] 2.2 worker-service：显式 `RequireRabbitMQ: true`
