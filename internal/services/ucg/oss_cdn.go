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

const videoSnapshotProcess = "video/snapshot,t_0"

// BuildVideoSnapshotURL 由视频 objectKey 拼装 OSS 首帧截帧 URL（通知封面快照）。
func BuildVideoSnapshotURL(objectKey string) string {
	cdn := BuildCdnURL(objectKey)
	if cdn == "" {
		return ""
	}
	return appendOssProcess(cdn, videoSnapshotProcess)
}

func appendOssProcess(cdn, process string) string {
	sep := "?"
	if strings.Contains(cdn, "?") {
		sep = "&"
	}
	return cdn + sep + "x-oss-process=" + process
}
