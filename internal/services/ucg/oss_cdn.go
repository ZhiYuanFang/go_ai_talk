package ucg

import (
	"context"
	"strings"
)

const imageThumbProcess = "image/auto-orient,1/resize,m_lfit,w_200/quality,q_90/format,jpg"

// BuildCdnURL 由 objectKey 拼装全分辨率 CDN URL。
func BuildCdnURL(objectKey string) string {
	cfg := LoadOSSConfig(context.Background())
	key := strings.TrimSpace(objectKey)
	if key == "" || cfg.CdnBaseURL == "" {
		return ""
	}
	return cfg.CdnBaseURL + "/" + strings.TrimPrefix(key, "/")
}

// BuildImageThumbnailURL 由 objectKey 拼装服务端 OSS 缩略图 URL（仅图片）。
func BuildImageThumbnailURL(objectKey string) string {
	cdn := BuildCdnURL(objectKey)
	if cdn == "" {
		return ""
	}
	return appendOssProcess(cdn, imageThumbProcess)
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
