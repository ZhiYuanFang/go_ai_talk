## ADDED Requirements

### Requirement: 系统 MUST 提供独立 push-service 进程

系统 MUST 以独立进程 `push-service` 承载厂商推送（APNs/HMS/MiPush）与设备 token 存储。该进程 MUST NOT 复用 `notify-service`。进程 MUST 使用独立配置文件（如 `config.push-service.yaml`）与可配置监听地址（默认建议 `:9808`）。数据库 MUST 使用同 MySQL 实例上的独立 schema `ai_voice_push`，token 表 MUST NOT 再作为 UCG 业务表写入权威。

#### Scenario: 进程可独立启动

- **WHEN** 运维仅启动 push-service 且数据库 `ai_voice_push` 可用
- **THEN** 服务 MUST 能接受注册与 internal 发送请求，且 MUST NOT 依赖 ucg-service 进程内存活才能连接厂商

### Requirement: Token 存储 MUST 在 ai_voice_push

push-service MUST 将设备推送 token 持久化于 `ai_voice_push` 中的设备表（实现表名可定为 `push_device`），并以 `(wx_id, device_key, channel)` 唯一约束支撑 upsert。channel MUST 限于 `apns`、`hms`、`mipush`。

#### Scenario: 同设备同通道覆盖 token

- **WHEN** 同一 `wxId`+`deviceKey`+`channel` 再次注册新 token
- **THEN** 系统 MUST 更新为新 token，MUST NOT 插入重复唯一键行

### Requirement: 部署清单 MUST 包含 push-service

CI ACR 工作流与 `docker-compose.microservices.yml`（及必要 overlay/env 示例）MUST 纳入 `push-service` 镜像构建/推送与本地编排；推送厂商相关环境变量 MUST 注入 push-service 容器，MUST NOT 再作为 ucg-service 推送实现的必需配置。

#### Scenario: ACR 可构建 push-service

- **WHEN** 工作流以全量或指定 `push-service` 触发构建
- **THEN** MUST 使用对应 Dockerfile 产出 push-service 镜像标签
