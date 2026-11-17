package gatewayapp

import (
	"context"
	"strings"

	"hello/internal/shared/assetpath"
	"hello/internal/shared/featurelogo"
)

// ApkObjectKeyPrefix 与 ucg ApkObjectKeyPrefix 保持一致（gateway 不 import services/ucg）。
const ApkObjectKeyPrefix = "apk/"

// LegacyApkURLPrefix 历史本地下载 path 前缀。
const LegacyApkURLPrefix = "/device/app/apk/"

// IsApkObjectKey 判断 download_url 是否为 OSS objectKey（新上传）。
func IsApkObjectKey(stored string) bool {
	key := strings.TrimPrefix(strings.TrimSpace(stored), "/")
	return strings.HasPrefix(key, ApkObjectKeyPrefix)
}

// IsLegacyApkPath 判断是否为历史 path-only 本地下载地址。
func IsLegacyApkPath(stored string) bool {
	p := strings.TrimSpace(stored)
	if strings.HasPrefix(strings.ToLower(p), "http://") || strings.HasPrefix(strings.ToLower(p), "https://") {
		p = assetpath.Normalize(p)
	}
	return strings.HasPrefix(p, LegacyApkURLPrefix)
}

// ResolveVersionDownloadURL 将库内 download_url 映射为 version/check / 管理列表对外值。
//
// 规则：
//   - 已是 http(s) → 原样返回（禁止剥 host）
//   - apk/ objectKey → CDN 绝对 URL（复用 UCG CDN 基址）
//   - 历史 /device/app/apk/ → path-only（App 可拼 gateway 基址）
func ResolveVersionDownloadURL(ctx context.Context, stored string) string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}
	lower := strings.ToLower(stored)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return stored
	}
	if IsApkObjectKey(stored) {
		return featurelogo.CdnURL(ctx, strings.TrimPrefix(stored, "/"))
	}
	return assetpath.Normalize(stored)
}

// ResolveVersionDownloadAbsoluteURL 官网/二维码用绝对下载地址。
//
// OSS/CDN 行直接返回绝对 URL；历史 path 行拼 publicBaseURL。
func ResolveVersionDownloadAbsoluteURL(ctx context.Context, stored, publicBaseURL string) string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}
	lower := strings.ToLower(stored)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return stored
	}
	if IsApkObjectKey(stored) {
		return featurelogo.CdnURL(ctx, strings.TrimPrefix(stored, "/"))
	}
	path := assetpath.Normalize(stored)
	if path == "" {
		return ""
	}
	if publicBaseURL == "" {
		return path
	}
	return strings.TrimRight(publicBaseURL, "/") + path
}
