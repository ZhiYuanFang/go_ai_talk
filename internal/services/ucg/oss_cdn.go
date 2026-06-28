package ucg

import (
	"context"
	"strings"

	"hello/internal/shared/mediacdn"
)

// BuildCdnURL 由 objectKey 拼装全分辨率 CDN URL。
func BuildCdnURL(objectKey string) string {
	cfg := LoadOSSConfig(context.Background())
	key := strings.TrimSpace(objectKey)
	if key == "" || cfg.CdnBaseURL == "" {
		return ""
	}
	return cfg.CdnBaseURL + "/" + strings.TrimPrefix(key, "/")
}

// BuildImageThumbnailURL 由原图 objectKey 拼装物理缩略图 CDN URL（path 为 stem_thumb.ext，无 x-oss-process）。
func BuildImageThumbnailURL(objectKey string) string {
	thumbKey := mediacdn.ThumbObjectKey(objectKey)
	if thumbKey == "" {
		return ""
	}
	return BuildCdnURL(thumbKey)
}

// BuildVideoThumbnailURL 由视频 objectKey 拼装物理首帧缩略图 CDN URL（path 为 stem_thumb.jpg，无 x-oss-process）。
func BuildVideoThumbnailURL(objectKey string) string {
	thumbKey := mediacdn.VideoThumbObjectKey(objectKey)
	if thumbKey == "" {
		return ""
	}
	return BuildCdnURL(thumbKey)
}
