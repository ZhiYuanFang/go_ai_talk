#!/usr/bin/env bash
# 检查业务/controller 层是否绕过 cachekit / redismsgkit 直连 Redis。
set -eu
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

fail=0

if rg 'g\.Redis\(\)' internal/services internal/controller --glob '*.go' 2>/dev/null; then
  echo "FAIL: 业务/controller 层存在 g.Redis() 直连" >&2
  fail=1
fi

if rg 'redis\.New(Client|ClusterClient)' internal/services internal/controller --glob '*.go' 2>/dev/null; then
  echo "FAIL: 业务/controller 层存在 go-redis 客户端创建" >&2
  fail=1
fi

if [ "$fail" -eq 0 ]; then
  echo "OK: Redis platform 访问合规"
fi
exit "$fail"
