// Package featurelogo 提供功能开通 logo 的 OSS objectKey 与 CDN URL 映射。
//
// 业务说明：功能视觉资源与事件 logo 同属 ucg OSS/CDN；库内存 feature/ 前缀 objectKey，
// HTTP 边界映射为 CDN 绝对 URL。本包供 cash 等调用方使用，禁止依赖 device 业务包。
package featurelogo

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

const defaultCdnBaseURL = "https://resorce.cuplay.top"

// ObjectKeyPrefix 功能 logo OSS 前缀（与 ucg 上传一致）。
const ObjectKeyPrefix = "feature/"

// CdnBaseURL 读取功能 logo CDN 根地址（与事件共用 UCG_OSS_CDN_BASE_URL / ucg.oss.cdnBaseUrl）。
func CdnBaseURL(ctx context.Context) string {
	if v := strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_OSS_CDN_BASE_URL")), "/"); v != "" {
		return v
	}
	if v := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.oss.cdnBaseUrl").String()), "/"); v != "" {
		return v
	}
	if v := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "device.eventLogoCdnBaseUrl").String()), "/"); v != "" {
		return v
	}
	return defaultCdnBaseURL
}

// NormalizeObjectKey 归一化库内 objectKey（去掉首尾空白与前导 /）。
func NormalizeObjectKey(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return strings.TrimPrefix(raw, "/")
}

// CdnURL 由 objectKey 拼装 CDN 绝对 URL；已是 http(s) 则原样返回。
func CdnURL(ctx context.Context, stored string) string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}
	lower := strings.ToLower(stored)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return stored
	}
	key := NormalizeObjectKey(stored)
	if key == "" {
		return ""
	}
	base := CdnBaseURL(ctx)
	if base == "" {
		return ""
	}
	return base + "/" + key
}

// StoredObjectKey 将 logo 规范为 DB 使用的 OSS objectKey；CDN 绝对 URL 会解析 path。
func StoredObjectKey(ctx context.Context, stored string) string {
	_ = ctx
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}
	lower := strings.ToLower(stored)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		u, err := url.Parse(stored)
		if err != nil || u.Path == "" || u.Path == "/" {
			return ""
		}
		return NormalizeObjectKey(u.Path)
	}
	return NormalizeObjectKey(stored)
}
