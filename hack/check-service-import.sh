#!/usr/bin/env bash
# 检查 internal/services 业务包之间是否互引（service-package-isolation）。
# 允许：同包（含 gatewayapp/usagestats→gatewayapp）、platform、contracts、aimodel、clients。
# 用法：bash hack/check-service-import.sh
set -eu
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

domains="cash voice device history ucg gatewayapp simuser mcpbridge appstatus"
fail=0

if ! command -v rg >/dev/null 2>&1; then
  echo "FAIL: 需要 ripgrep (rg)" >&2
  exit 1
fi

for src in $domains; do
  src_dir="internal/services/$src"
  if [ ! -d "$src_dir" ]; then
    continue
  fi
  for dst in $domains; do
    if [ "$src" = "$dst" ]; then
      continue
    fi
    hits="$(rg -n --glob '*.go' "\"hello/internal/services/${dst}\"" "$src_dir" 2>/dev/null || true)"
    if [ -n "$hits" ]; then
      echo "FAIL: services/$src 禁止 import services/$dst" >&2
      echo "$hits" >&2
      fail=1
    fi
  done
done

if [ "$fail" -eq 0 ]; then
  echo "OK: services 业务包互引检查通过"
fi
exit "$fail"
