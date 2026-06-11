package eventlogo

import (
	"context"
	"net/url"
	"os"
	"strings"

	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

const defaultCdnBaseURL = "https://resorce.cuplay.top"

// ObjectKeyPrefix 事件 logo OSS 前缀（与 ucg 上传、迁移脚本一致）。
const ObjectKeyPrefix = "event/"

// CdnBaseURL 读取事件 logo CDN 根地址（环境变量优先，便于 prod/test 注入）。
func CdnBaseURL(ctx context.Context) string {
	if v := strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_OSS_CDN_BASE_URL")), "/"); v != "" {
		return v
	}
	if v := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "device.eventLogoCdnBaseUrl").String()), "/"); v != "" {
		return v
	}
	if v := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.oss.cdnBaseUrl").String()), "/"); v != "" {
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

// StoredObjectKey 将 logo 规范为 DB/Redis 使用的 OSS objectKey；CDN 绝对 URL 会解析 path 并去掉前导 /。
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

// MapEventsLogoStored 写 Redis 前将事件列表 logo 统一为 objectKey，避免与 device 重建缓存的格式混写。
func MapEventsLogoStored(ctx context.Context, rows []entity.Event) []entity.Event {
	if len(rows) == 0 {
		return rows
	}
	out := make([]entity.Event, len(rows))
	copy(out, rows)
	for i := range out {
		out[i].Logo = StoredObjectKey(ctx, out[i].Logo)
	}
	return out
}

// MapEventsLogoCdn 将事件列表 logo 字段映射为 CDN URL（供 HTTP 响应序列化）。
func MapEventsLogoCdn(ctx context.Context, rows []entity.Event) []entity.Event {
	if len(rows) == 0 {
		return rows
	}
	out := make([]entity.Event, len(rows))
	copy(out, rows)
	for i := range out {
		out[i].Logo = CdnURL(ctx, out[i].Logo)
	}
	return out
}
