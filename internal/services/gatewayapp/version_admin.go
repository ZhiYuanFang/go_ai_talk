package gatewayapp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	// VersionAdminSessionCookieName 版本管理页登录态 Cookie 名。
	VersionAdminSessionCookieName = "gw_ver_admin"
	// redisKeyVersionAdminSessionPrefix Redis 中管理员会话 key 前缀。
	redisKeyVersionAdminSessionPrefix = "gw:app:veradmin:sess:"
	// RedisKeyAppVersionLatestCache 与 gateway 版本检查缓存键一致，写库后须删除以便立即读到新行。
	RedisKeyAppVersionLatestCache = "gw:app:version:latest"
)

// AppVersionLatestCacheKey 按 APK 存储目录区分 prod/test，避免双栈共用 Redis 时版本缓存串环境。
func AppVersionLatestCacheKey(ctx context.Context) string {
	dir := strings.TrimSpace(ApkStorageDir(ctx))
	if dir == "" {
		return RedisKeyAppVersionLatestCache
	}
	return RedisKeyAppVersionLatestCache + ":" + dir
}

// VersionAdminPassword 管理员口令：优先环境变量 GATEWAY_APP_VERSION_ADMIN_PASSWORD，否则读配置 gatewayApp.versionAdmin.password（生产勿将真口令写入仓库 yaml）。
func VersionAdminPassword(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("GATEWAY_APP_VERSION_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.versionAdmin.password").String())
}

// PublicBaseURL 对外访问 gateway-app 的基址（用于拼接 APK 绝对下载 URL），须含 scheme，建议不含末尾斜杠。
func PublicBaseURL(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("GATEWAY_APP_PUBLIC_BASE_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.publicBaseUrl").String()), "/")
}

// ApkStorageDir APK 落盘目录，默认 /apk/ai_talk/；可通过 gatewayApp.apkStorageDir 或环境变量 GATEWAY_APP_APK_STORAGE_DIR 覆盖。
func ApkStorageDir(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("GATEWAY_APP_APK_STORAGE_DIR")); v != "" {
		v = strings.TrimSpace(v)
		if !strings.HasSuffix(v, string(os.PathSeparator)) && v != "" {
			v += string(os.PathSeparator)
		}
		return v
	}
	s := strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.apkStorageDir").String())
	if s == "" {
		return "/apk/ai_talk/"
	}
	if !strings.HasSuffix(s, "/") && !strings.HasSuffix(s, "\\") {
		s += "/"
	}
	return s
}

// ApkMaxBytes 单文件大小上限，默认 200MB。
func ApkMaxBytes(ctx context.Context) int64 {
	if v := strings.TrimSpace(os.Getenv("GATEWAY_APP_APK_MAX_BYTES")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	n := g.Cfg().MustGet(ctx, "gatewayApp.apkMaxBytes").Int64()
	if n <= 0 {
		return 200 << 20
	}
	return n
}

// VersionAdminSessionTTL 管理会话 TTL，默认 8 小时。
func VersionAdminSessionTTL(ctx context.Context) time.Duration {
	sec := g.Cfg().MustGet(ctx, "gatewayApp.versionAdmin.sessionTtlSeconds").Int64()
	if sec <= 0 {
		sec = 8 * 3600
	}
	return time.Duration(sec) * time.Second
}

// VersionAdminCookieSecure 是否对管理 Cookie 设置 Secure（HTTPS）；可由 gatewayApp.versionAdmin.cookieSecure 或环境 GATEWAY_APP_VERSION_ADMIN_COOKIE_SECURE=1 开启。
func VersionAdminCookieSecure(ctx context.Context) bool {
	if strings.TrimSpace(os.Getenv("GATEWAY_APP_VERSION_ADMIN_COOKIE_SECURE")) == "1" {
		return true
	}
	return g.Cfg().MustGet(ctx, "gatewayApp.versionAdmin.cookieSecure").Bool()
}

// NewVersionAdminSessionID 生成随机会话 ID（32 hex）。
func NewVersionAdminSessionID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// RedisSessionKey 会话在 Redis 中的完整 key。
func RedisSessionKey(sessionID string) string {
	return redisKeyVersionAdminSessionPrefix + strings.TrimSpace(sessionID)
}

// InvalidateAppVersionLatestCache 删除版本检查用的最新行缓存，避免上传后仍返回旧 downloadUrl。
func InvalidateAppVersionLatestCache(ctx context.Context) {
	key := AppVersionLatestCacheKey(ctx)
	if _, err := g.Redis().Do(ctx, "DEL", key); err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 删除版本缓存失败 key=%s err=%v", key, err)
	}
}
