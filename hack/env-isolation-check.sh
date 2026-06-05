#!/usr/bin/env bash
# 测试/正式环境隔离只读验收（Redis + MySQL + 对外入口）。
# 用于排查：测试服偶现正式数据、注销成功但观测库无变化等环境串线问题。
#
# 用法（在仓库根目录）：
#   chmod +x hack/env-isolation-check.sh
#   ./hack/env-isolation-check.sh
#
# 出问题时定点排查（可选）：
#   export INCIDENT_WX_ID=42
#   export INCIDENT_DEVICE_NO=ABCDEF
#   export MYSQL_CLI_USER=root MYSQL_CLI_PASS='***' MYSQL_HOST=120.55.50.105
#   ./hack/env-isolation-check.sh

set -uo pipefail

PASS=0
WARN=0
FAIL=0

COMPOSE_REDIS_CLUSTER="${COMPOSE_REDIS_CLUSTER:-manifest/docker/docker-compose.redis-cluster.yml}"
TEST_GATEWAY_URL="${TEST_GATEWAY_URL:-https://test.pangbao.cuplay.top:9702}"
PROD_GATEWAY_URL="${PROD_GATEWAY_URL:-https://www.pangbao.cuplay.top:9702}"

ok()   { echo "  [PASS] $*"; PASS=$((PASS + 1)); }
warn() { echo "  [WARN] $*"; WARN=$((WARN + 1)); }
bad()  { echo "  [FAIL] $*"; FAIL=$((FAIL + 1)); }
hr()   { echo ""; echo "════════════════════════════════════════════════════════"; echo " $*"; echo "════════════════════════════════════════════════════════"; }

container_running() {
  docker ps --format '{{.Names}}' | grep -qx "$1"
}

prod_redis_cli() {
  # 通过 compose project 执行，避免宿主机容器名随 project 变化。
  docker compose -f "$COMPOSE_REDIS_CLUSTER" exec -T redis-node-1 redis-cli -p 7001 "$@" 2>/dev/null
}

prod_redis_scan() {
  docker compose -f "$COMPOSE_REDIS_CLUSTER" exec -T redis-node-1 redis-cli -p 7001 --scan --pattern "$1" 2>/dev/null
}

# ── 0. 前置 ──────────────────────────────────────────────
hr "0. 前置检查"
if ! command -v docker >/dev/null 2>&1; then
  bad "docker 未安装"
  exit 1
fi
ok "docker 可用"

if [[ ! -f "manifest/docker/.env.test" ]]; then
  warn "当前目录可能不是仓库根（找不到 manifest/docker/.env.test）"
else
  ok "仓库根目录 manifest/docker/.env.test 存在"
fi

# ── 1. 容器与网络 ────────────────────────────────────────
hr "1. 容器运行状态"
TEST_CONTAINERS=(
  go-ai-talk-gateway-app-test
  go-ai-talk-device-service-test
  go-ai-talk-history-service-test
  go-ai-talk-voice-service-test
  go-ai-talk-worker-test
  go-ai-talk-redis-test
)
PROD_CONTAINERS=(
  go-ai-talk-gateway-app
  go-ai-talk-device-service
  go-ai-talk-history-service
)

for c in "${TEST_CONTAINERS[@]}"; do
  if container_running "$c"; then ok "测试容器运行中: $c"
  else warn "测试容器未运行: $c"; fi
done
for c in "${PROD_CONTAINERS[@]}"; do
  if container_running "$c"; then ok "正式容器运行中: $c"
  else warn "正式容器未运行: $c"; fi
done

if prod_redis_cli PING | grep -q PONG; then ok "正式 Redis Cluster 可访问（compose redis-node-1:7001）"
else warn "正式 Redis Cluster 不可访问（可能未启动或 compose 路径不对）"; fi

hr "1b. Docker 网络"
for net in go-ai-talk-test-net go-ai-talk-net; do
  if docker network inspect "$net" >/dev/null 2>&1; then ok "网络存在: $net"
  else warn "网络不存在: $net"; fi
done

# ── 2. 测试栈环境变量 ────────────────────────────────────
hr "2. 测试微服务环境变量（须 _test 库 + redis-test:6379）"

check_test_env() {
  local container="$1"
  local db_var="$2"

  if ! container_running "$container"; then
    warn "$container 未运行，跳过"
    return
  fi

  echo "  --- $container ---"
  local redis_addr db_link jwt_sec
  redis_addr=$(docker exec "$container" printenv GF_REDIS_DEFAULT_ADDRESS 2>/dev/null || true)
  db_link=$(docker exec "$container" printenv "$db_var" 2>/dev/null || true)
  jwt_sec=$(docker exec "$container" printenv GATEWAY_APP_JWT_SECRET 2>/dev/null || true)

  echo "      GF_REDIS_DEFAULT_ADDRESS=$redis_addr"
  echo "      $db_var=$db_link"
  [[ -n "$jwt_sec" ]] && echo "      GATEWAY_APP_JWT_SECRET=${jwt_sec:0:12}..."

  if [[ "$redis_addr" == "redis-test:6379" ]]; then
    ok "$container Redis 地址正确"
  elif [[ -z "$redis_addr" ]]; then
    bad "$container GF_REDIS_DEFAULT_ADDRESS 为空"
  elif [[ "$redis_addr" == *"redis-node"* || "$redis_addr" == *"7001"* ]]; then
    bad "$container 疑似连正式 Redis: $redis_addr"
  else
    warn "$container Redis 地址非预期: $redis_addr"
  fi

  if [[ "$db_link" == *"_test"* ]]; then
    ok "$container 库名含 _test"
  elif [[ -n "$db_link" ]]; then
    bad "$container 库名可能为正式库: $db_link"
  else
    warn "$container $db_var 未设置"
  fi
}

check_test_env go-ai-talk-gateway-app-test APP_DB_LINK
check_test_env go-ai-talk-device-service-test DEVICE_DB_LINK
check_test_env go-ai-talk-history-service-test HISTORY_DB_LINK
check_test_env go-ai-talk-voice-service-test VOICE_DB_LINK
check_test_env go-ai-talk-worker-test WORKER_OUTBOX_DB_LINK

if container_running go-ai-talk-device-service-test && container_running go-ai-talk-gateway-app-test; then
  dsec=$(docker exec go-ai-talk-device-service-test printenv DEVICE_GATEWAY_INTERNAL_SECRET 2>/dev/null || true)
  gsec=$(docker exec go-ai-talk-gateway-app-test printenv DEVICE_GATEWAY_INTERNAL_SECRET 2>/dev/null || true)
  echo "  --- 内部密钥一致性 ---"
  echo "      device DEVICE_GATEWAY_INTERNAL_SECRET=${dsec:0:20}..."
  echo "      gateway-app DEVICE_GATEWAY_INTERNAL_SECRET=${gsec:0:20}..."
  if [[ -n "$dsec" && "$dsec" == "$gsec" ]]; then ok "gateway-app 与 device 内部密钥一致"
  else bad "gateway-app 与 device 内部密钥不一致"; fi
fi

# ── 3. 测试容器 DNS ──────────────────────────────────────
hr "3. 测试容器 DNS 解析"
if container_running go-ai-talk-device-service-test; then
  for host in redis-test redis-node-1 redis-node-2 redis-node-3; do
    ip=$(docker exec go-ai-talk-device-service-test getent hosts "$host" 2>/dev/null | awk '{print $1}' | head -1 || true)
    if [[ -n "$ip" ]]; then
      if [[ "$host" == "redis-test" ]]; then ok "device-test 可解析 $host → $ip"
      else bad "device-test 能解析正式 Redis 主机 $host → $ip"; fi
    else
      if [[ "$host" == "redis-test" ]]; then bad "device-test 无法解析 redis-test"
      else ok "device-test 无法解析 $host（符合隔离预期）"; fi
    fi
  done
fi

# ── 4. 启动日志 ──────────────────────────────────────────
hr "4. 启动日志中的 database / redis 配置行"
for c in go-ai-talk-gateway-app-test go-ai-talk-device-service-test go-ai-talk-history-service-test; do
  if ! container_running "$c"; then continue; fi
  echo "  --- $c ---"
  docker logs "$c" 2>&1 | grep -E 'database\.(default|app) 已用|redis\.default 已用' | tail -3 | sed 's/^/      /' || true
  if docker logs "$c" 2>&1 | grep -q 'redis.default 已用'; then
    line=$(docker logs "$c" 2>&1 | grep 'redis.default 已用' | tail -1)
    if echo "$line" | grep -q 'redis-test:6379'; then ok "$c 日志确认连测试 Redis"
    elif echo "$line" | grep -q 'redis-node'; then bad "$c 日志显示连正式 Redis Cluster"
    else warn "$c redis 日志: $line"; fi
  else
    warn "$c 启动日志无 redis.default 行"
  fi
done

# ── 5. 对外 HTTPS 入口 ───────────────────────────────────
hr "5. 对外 HTTPS 入口（客户端视角）"

probe_site_home() {
  local label="$1" url="$2" expect_db="$3"
  echo "  --- $label: $url ---"
  body=$(curl -sk --connect-timeout 8 --max-time 15 "$url/device/app/api/site/home" 2>/dev/null || true)
  if [[ -z "$body" ]]; then
    bad "$label site/home 无响应"
    return
  fi
  db=$(echo "$body" | grep -o '"appDatabase":"[^"]*"' | head -1 | cut -d'"' -f4)
  pub=$(echo "$body" | grep -o '"publicBaseUrl":"[^"]*"' | head -1 | cut -d'"' -f4)
  echo "      appDatabase=$db"
  echo "      publicBaseUrl=$pub"
  if [[ "$db" == "$expect_db" ]]; then ok "$label appDatabase 正确 ($expect_db)"
  elif [[ -n "$db" ]]; then bad "$label appDatabase=$db，期望 $expect_db"
  else warn "$label 未解析到 appDatabase"; fi
}

probe_site_home "测试" "$TEST_GATEWAY_URL" "ai_voice_app_test"
probe_site_home "正式" "$PROD_GATEWAY_URL" "ai_voice_app"

# ── 6. Redis 实例指纹 ────────────────────────────────────
hr "6. Redis 实例指纹（run_id 相同则共用同一实例）"

test_run_id=""
prod_run_id=""
if container_running go-ai-talk-redis-test; then
  test_run_id=$(docker exec go-ai-talk-redis-test redis-cli INFO server 2>/dev/null | grep '^run_id:' | cut -d: -f2 | tr -d '\r')
  test_keys=$(docker exec go-ai-talk-redis-test redis-cli DBSIZE 2>/dev/null | awk '{print $2}')
  echo "  测试 Redis (go-ai-talk-redis-test): run_id=$test_run_id dbsize=$test_keys"
fi
if prod_redis_cli PING | grep -q PONG; then
  prod_run_id=$(prod_redis_cli INFO server | grep '^run_id:' | cut -d: -f2 | tr -d '\r')
  prod_keys=$(prod_redis_cli DBSIZE | awk '{print $2}')
  echo "  正式 Redis (redis-node-1:7001): run_id=$prod_run_id dbsize=$prod_keys"
fi
if [[ -n "$test_run_id" && -n "$prod_run_id" && "$test_run_id" == "$prod_run_id" ]]; then
  bad "测试与正式 Redis run_id 相同 — 共用同一实例"
else
  ok "测试与正式 Redis run_id 不同（或仅一侧可测）"
fi

# ── 7. Redis 关键 Key 抽样 ───────────────────────────────
hr "7. Redis 关键 Key 抽样（只读）"

sample_test_redis_keys() {
  if ! container_running go-ai-talk-redis-test; then
    warn "测试 Redis 未运行，跳过"
    return
  fi
  echo "  --- 测试 Redis ---"
  for pattern in 'dev:wx:id2dev:*' 'dev:wx:id2union:*' 'history:record:list:*' 'gw:app:rt:*' 'voice:session:*'; do
    count=$(docker exec go-ai-talk-redis-test redis-cli --scan --pattern "$pattern" 2>/dev/null | wc -l | tr -d ' ')
    echo "      $pattern → ${count} 个 key"
  done
  echo "      抽样 dev:wx:id2dev（最多 3 条）:"
  docker exec go-ai-talk-redis-test redis-cli --scan --pattern 'dev:wx:id2dev:*' 2>/dev/null | head -3 | while read -r k; do
    [[ -z "$k" ]] && continue
    v=$(docker exec go-ai-talk-redis-test redis-cli GET "$k" 2>/dev/null)
    ttl=$(docker exec go-ai-talk-redis-test redis-cli TTL "$k" 2>/dev/null)
    echo "        $k = $v (ttl=${ttl}s)"
  done
}

sample_prod_redis_keys() {
  if ! prod_redis_cli PING | grep -q PONG; then
    warn "正式 Redis 不可访问，跳过"
    return
  fi
  echo "  --- 正式 Redis (node-1) ---"
  for pattern in 'dev:wx:id2dev:*' 'dev:wx:id2union:*' 'history:record:list:*' 'gw:app:rt:*' 'voice:session:*'; do
    count=$(prod_redis_scan "$pattern" | wc -l | tr -d ' ')
    echo "      $pattern → ${count} 个 key"
  done
}

sample_test_redis_keys
sample_prod_redis_keys

# ── 8. 测试 Redis PING ───────────────────────────────────
hr "8. 测试 Redis 连通"
if container_running go-ai-talk-redis-test; then
  if docker exec go-ai-talk-redis-test redis-cli PING 2>/dev/null | grep -q PONG; then ok "测试 Redis PING 正常"
  else bad "测试 Redis PING 失败"; fi
fi

# ── 9. 定点排查（可选）──────────────────────────────────
hr "9. 定点排查（export INCIDENT_* / MYSQL_CLI_* 后生效）"

if [[ -n "${INCIDENT_WX_ID:-}" ]]; then
  echo "  wxId=$INCIDENT_WX_ID"
  if container_running go-ai-talk-redis-test; then
    for k in "dev:wx:id2dev:${INCIDENT_WX_ID}" "dev:wx:id2union:${INCIDENT_WX_ID}"; do
      v=$(docker exec go-ai-talk-redis-test redis-cli GET "$k" 2>/dev/null || true)
      ttl=$(docker exec go-ai-talk-redis-test redis-cli TTL "$k" 2>/dev/null || true)
      echo "      [test redis] $k = ${v:-<nil>} ttl=${ttl}s"
    done
  fi
  if prod_redis_cli PING | grep -q PONG; then
    for k in "dev:wx:id2dev:${INCIDENT_WX_ID}" "dev:wx:id2union:${INCIDENT_WX_ID}"; do
      v=$(prod_redis_cli GET "$k" 2>/dev/null || true)
      ttl=$(prod_redis_cli TTL "$k" 2>/dev/null || true)
      echo "      [prod redis] $k = ${v:-<nil>} ttl=${ttl}s"
    done
  fi
  if command -v mysql >/dev/null 2>&1 && [[ -n "${MYSQL_CLI_USER:-}" ]]; then
    echo "  MySQL 对照:"
    mysql -h"${MYSQL_HOST:-120.55.50.105}" -u"$MYSQL_CLI_USER" -p"${MYSQL_CLI_PASS:-}" -N -e \
      "SELECT 'test', id, unionid, device_no FROM ai_voice_device_test.wx WHERE id=${INCIDENT_WX_ID};
       SELECT 'prod', id, unionid, device_no FROM ai_voice_device.wx WHERE id=${INCIDENT_WX_ID};" 2>/dev/null \
      | sed 's/^/      /' || warn "MySQL 查询失败"
  else
    echo "      跳过 MySQL（设置 MYSQL_CLI_USER / MYSQL_CLI_PASS / MYSQL_HOST）"
  fi
fi

if [[ -n "${INCIDENT_DEVICE_NO:-}" ]]; then
  echo "  deviceNo=$INCIDENT_DEVICE_NO"
  dn_lower=$(echo "$INCIDENT_DEVICE_NO" | tr '[:upper:]' '[:lower:]')
  if container_running go-ai-talk-redis-test; then
    for k in "history:record:list:${INCIDENT_DEVICE_NO}" "history:record:list:${dn_lower}"; do
      exists=$(docker exec go-ai-talk-redis-test redis-cli EXISTS "$k" 2>/dev/null || echo 0)
      echo "      [test redis] EXISTS $k = $exists"
    done
  fi
fi

# ── 10. 汇总 ─────────────────────────────────────────────
hr "10. 汇总"
echo "  PASS=$PASS  WARN=$WARN  FAIL=$FAIL"
echo ""
if [[ $FAIL -gt 0 ]]; then
  echo "  结论：存在明确串线风险，优先处理 FAIL 项（Redis 地址、库名、run_id、appDatabase）。"
elif [[ $WARN -gt 0 ]]; then
  echo "  结论：无硬性 FAIL，但有 WARN；结合出问题时 site/home 与 JWT wxId 继续定位。"
else
  echo "  结论：基础设施隔离正常；若仍偶现，重点查客户端 base URL / token 缓存。"
fi
echo ""
echo "  出问题时追加："
echo "    INCIDENT_WX_ID=<JWT sub> INCIDENT_DEVICE_NO=<device_no> \\"
echo "    MYSQL_CLI_USER=root MYSQL_CLI_PASS='***' ./hack/env-isolation-check.sh"

if [[ $FAIL -gt 0 ]]; then exit 1; fi
exit 0
