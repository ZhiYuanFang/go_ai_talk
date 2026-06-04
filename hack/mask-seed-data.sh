#!/usr/bin/env sh
# 从生产 ai_voice_* 库导出、脱敏并导入 ai_voice_*_test 测试库。
# 同步事件 logo 静态文件至 /ai_talk_images_test（可选）。
#
# 用法（在仓库根目录，需本机 mysql/mysqldump 客户端）：
#   chmod +x hack/mask-seed-data.sh
#   MYSQL_HOST=127.0.0.1 MYSQL_USER=root MYSQL_PASS='***' ./hack/mask-seed-data.sh
#
# 可选环境变量：
#   MYSQL_HOST       默认 127.0.0.1
#   MYSQL_PORT       默认 3306
#   MYSQL_USER       默认 root
#   MYSQL_PASS       必填（或已配置 ~/.my.cnf）
#   DUMP_DIR         默认 ./tmp/seed-dump
#   SKIP_STATIC_SYNC=1  跳过 rsync/cp logo 至 /ai_talk_images_test
#
# 警告：导入会 DROP/覆盖 _test 库中已有数据；执行前确认 DSN 库名均带 _test 后缀。

set -eu

MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASS="${MYSQL_PASS:-Fang930927}"
DUMP_DIR="${DUMP_DIR:-./tmp/seed-dump}"

if [ -z "${MYSQL_PASS}" ] && [ ! -f "${HOME}/.my.cnf" ]; then
  echo "请设置 MYSQL_PASS 或配置 ~/.my.cnf" >&2
  exit 1
fi

mysql_base() {
  if [ -n "${MYSQL_PASS}" ]; then
    mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p"${MYSQL_PASS}" "$@"
  else
    mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" "$@"
  fi
}

mysqldump_base() {
  if [ -n "${MYSQL_PASS}" ]; then
    mysqldump -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p"${MYSQL_PASS}" "$@"
  else
    mysqldump -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" "$@"
  fi
}

PROD_DBS="ai_voice_history ai_voice_device ai_voice_voice ai_voice_worker ai_voice_app ai_voice_ucg"

mkdir -p "${DUMP_DIR}/raw" "${DUMP_DIR}/masked"

echo "==> 导出生产库..."
for db in ${PROD_DBS}; do
  echo "  dump ${db}"
  mysqldump_base --single-transaction --routines --triggers "${db}" \
    > "${DUMP_DIR}/raw/${db}.sql"
done

echo "==> 脱敏（复制 raw → masked 并替换库名）..."
for db in ${PROD_DBS}; do
  test_db="${db}_test"
  sed "s/\`${db}\`/\`${test_db}\`/g; s/USE \`${db}\`/USE \`${test_db}\`/g" \
    "${DUMP_DIR}/raw/${db}.sql" > "${DUMP_DIR}/masked/${test_db}.sql"
done

echo "==> 创建测试库（若不存在）..."
for db in ${PROD_DBS}; do
  test_db="${db}_test"
  mysql_base -e "CREATE DATABASE IF NOT EXISTS \`${test_db}\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
done

echo "==> 导入测试库..."
for db in ${PROD_DBS}; do
  test_db="${db}_test"
  echo "  import ${test_db}"
  mysql_base -e "DROP DATABASE IF EXISTS \`${test_db}\`; CREATE DATABASE \`${test_db}\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
  mysql_base "${test_db}" < "${DUMP_DIR}/masked/${test_db}.sql"
done

echo "==> 表级脱敏（device / wx 等，按实际表结构尽力覆盖）..."
mysql_base ai_voice_device_test 2>/dev/null <<'EOSQL' || true
UPDATE user SET device_no = CONCAT('T_', device_no) WHERE device_no IS NOT NULL AND device_no != '' AND device_no NOT LIKE 'T_%';
UPDATE user SET baby_name = CONCAT('测试宝宝_', id) WHERE baby_name IS NOT NULL AND baby_name != '';
UPDATE wx SET unionid = CONCAT('test_unionid_', id) WHERE unionid IS NOT NULL AND unionid != '';
UPDATE wx SET account = CONCAT('test_account_', id) WHERE account IS NOT NULL AND account != '';
UPDATE wx SET device_no = CONCAT('T_', device_no) WHERE device_no IS NOT NULL AND device_no != '' AND device_no NOT LIKE 'T_%';
UPDATE wx SET password = '$2a$10$testplaceholderhashnotforlogin' WHERE password IS NOT NULL AND password != '';
EOSQL

echo "==> 静态资源同步（生产 logo → 测试目录）..."
if [ "${SKIP_STATIC_SYNC:-0}" != "1" ] && [ -d /ai_talk_images ]; then
  sudo mkdir -p /ai_talk_images_test
  sudo rsync -a --delete /ai_talk_images/ /ai_talk_images_test/ 2>/dev/null \
    || sudo cp -a /ai_talk_images/. /ai_talk_images_test/ 2>/dev/null \
    || echo "  跳过静态同步（无权限或目录不存在，请手工 cp/rsync）"
else
  echo "  跳过（SKIP_STATIC_SYNC=1 或 /ai_talk_images 不存在）"
fi

echo ""
echo "脱敏种子导入完成。请 force-recreate 测试栈中依赖 DB/Redis 缓存的服务。"
echo "验收：mysql -e 'SHOW DATABASES LIKE \"ai_voice%test\"' ; ls /ai_talk_images_test"
