# 启用 rabbitmq_delayed_message_exchange 的 RabbitMQ 镜像（基于官方 management）。
# 构建：docker compose -f manifest/docker/docker-compose.rabbitmq.yml build
FROM rabbitmq:3.13-management-alpine

ARG DELAYED_PLUGIN_VERSION=3.13.0

# 社区延时交换机插件；离线 enable 后随容器启动生效。
RUN apk add --no-cache curl \
	&& curl -fsSL -o /opt/rabbitmq/plugins/rabbitmq_delayed_message_exchange-${DELAYED_PLUGIN_VERSION}.ez \
		"https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/v${DELAYED_PLUGIN_VERSION}/rabbitmq_delayed_message_exchange-${DELAYED_PLUGIN_VERSION}.ez" \
	&& rabbitmq-plugins enable --offline rabbitmq_delayed_message_exchange
