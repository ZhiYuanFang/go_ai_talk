package gatewayapp

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	// RedisKeyAppVersionLatestCache 与 gateway 版本检查缓存键一致，写库后须删除以便立即读到新行。
	RedisKeyAppVersionLatestCache = "gw:app:version:latest"
)

// AppVersionLatestCacheKey 按对外基址或 APP 库名区分 prod/test。
// 注意：test 卷将 /apk/ai_talk_test 挂载到容器内 /apk/ai_talk，不可用 ApkStorageDir 区分环境。
func AppVersionLatestCacheKey(ctx context.Context) string {
	if base := strings.TrimSpace(PublicBaseURL(ctx)); base != "" {
		return RedisKeyAppVersionLatestCache + ":" + base
	}
	if link := strings.TrimSpace(os.Getenv("APP_DB_LINK")); link != "" {
		if i := strings.LastIndex(link, "/"); i >= 0 && i < len(link)-1 {
			if dbName := strings.TrimSpace(link[i+1:]); dbName != "" {
				return RedisKeyAppVersionLatestCache + ":" + dbName
			}
		}
	}
	return RedisKeyAppVersionLatestCache
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

// InvalidateAppVersionLatestCache 删除版本检查用的最新行缓存，避免上传后仍返回旧 downloadUrl。
func InvalidateAppVersionLatestCache(ctx context.Context) {
	keys := []string{
		AppVersionLatestCacheKey(ctx),
		RedisKeyAppVersionLatestCache,
	}
	if dir := strings.TrimSpace(ApkStorageDir(ctx)); dir != "" {
		keys = append(keys, RedisKeyAppVersionLatestCache+":"+dir)
	}
	for _, key := range keys {
		if _, err := g.Redis().Do(ctx, "DEL", key); err != nil {
			glog.Warningf(ctx, "[gateway-app-version-admin] 删除版本缓存失败 key=%s err=%v", key, err)
		}
	}
}
