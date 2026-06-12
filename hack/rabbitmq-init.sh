#!/usr/bin/env sh
# RabbitMQ 基线：声明 voice.events（topic）及队列与绑定。与 hack/rabbitmq-init.ps1 语义一致。
# 用法（在仓库根目录）：
#   chmod +x hack/rabbitmq-init.sh
#   ./hack/rabbitmq-init.sh
# 可选环境变量：
#   COMPOSE_FILE=manifest/docker/docker-compose.rabbitmq.yml
#   RABBIT_API_BASE=http://127.0.0.1:15672/api   （宿主机执行时用映射端口；在容器内可改为 http://rabbitmq:15672/api）
#   RABBITMQ_USER=guest  RABBITMQ_PASS=guest
#   SKIP_UP=1  仅声明拓扑，不执行 docker compose up

set -eu

COMPOSE_FILE="${COMPOSE_FILE:-manifest/docker/docker-compose.rabbitmq.yml}"
API_BASE="${RABBIT_API_BASE:-http://127.0.0.1:15672/api}"
USER="${RABBITMQ_USER:-guest}"
PASS="${RABBITMQ_PASS:-guest}"

rabbit_put() {
  path="$1"
  body="$2"
  curl -sfS -u "${USER}:${PASS}" -H "Content-Type: application/json" \
    -X PUT "${API_BASE%/}${path}" -d "${body}"
}

rabbit_post() {
  path="$1"
  body="$2"
  curl -sfS -u "${USER}:${PASS}" -H "Content-Type: application/json" \
    -X POST "${API_BASE%/}${path}" -d "${body}"
}

wait_api() {
  i=0
  while [ "$i" -lt 30 ]; do
    if curl -sfS -u "${USER}:${PASS}" -o /dev/null "${API_BASE%/}/overview"; then
      return 0
    fi
    i=$((i + 1))
    sleep 1
  done
  echo "RabbitMQ management API not ready: ${API_BASE}" >&2
  exit 1
}

if [ "${SKIP_UP:-0}" != "1" ]; then
  echo "Starting RabbitMQ..."
  docker compose -f "${COMPOSE_FILE}" up -d
  wait_api
else
  wait_api
fi

echo "Declaring exchange and queues..."

rabbit_put "/exchanges/%2F/voice.events" \
  '{"type":"topic","durable":true,"auto_delete":false}'

# 队列名与路由键须与 internal/shared/mq 及 worker 消费约定一致。
declare_queue_bind() {
  name="$1"
  rk="$2"
  rabbit_put "/queues/%2F/${name}" '{"durable":true,"auto_delete":false,"arguments":{}}'
  rabbit_post "/bindings/%2F/e/voice.events/q/${name}" \
    "{\"routing_key\":\"${rk}\",\"arguments\":{}}"
}

declare_queue_bind "voice.task.requested.q" "voice.task.requested"
declare_queue_bind "voice.task.completed.q" "voice.task.completed"
declare_queue_bind "voice.task.failed.q" "voice.task.failed"
declare_queue_bind "notify.events.q" "notify.*"
declare_queue_bind "history.events.q" "history.#"
declare_queue_bind "ucg.post.created.q" "ucg.post.created"
declare_queue_bind "ucg.comment.created.q" "ucg.comment.created"
declare_queue_bind "ucg.profile.patch.submitted.q" "ucg.profile.patch.submitted"
declare_queue_bind "ucg.chat.msg.created.q" "ucg.chat.msg.created"
declare_queue_bind "ucg.recommend.score.q" "ucg.post.published"
declare_queue_bind "ucg.recommend.score.q" "ucg.post.unpublished"
declare_queue_bind "ucg.recommend.score.q" "ucg.post.liked"
declare_queue_bind "ucg.recommend.score.q" "ucg.post.unliked"
declare_queue_bind "ucg.recommend.score.q" "ucg.comment.published"
declare_queue_bind "ucg.recommend.score.q" "ucg.comment.removed"

echo ""
echo "RabbitMQ baseline initialized."
echo "Exchange: voice.events (topic)"
echo "Queues:"
echo " - voice.task.requested.q <= voice.task.requested"
echo " - voice.task.completed.q <= voice.task.completed"
echo " - voice.task.failed.q <= voice.task.failed"
echo " - notify.events.q <= notify.*"
echo " - history.events.q <= history.#"
echo " - ucg.post.created.q <= ucg.post.created"
echo " - ucg.comment.created.q <= ucg.comment.created"
echo " - ucg.profile.patch.submitted.q <= ucg.profile.patch.submitted"
echo " - ucg.chat.msg.created.q <= ucg.chat.msg.created"
echo " - ucg.recommend.score.q <= ucg.post.published|unpublished|liked|unliked|comment.published|comment.removed"
echo ""
echo "Verify: open management UI (e.g. http://127.0.0.1:15672 )"
