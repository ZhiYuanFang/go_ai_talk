package mediacdn

import (
	"path/filepath"
	"strings"
)

// ImageThumbSuffix 图片缩略图 objectKey 在扩展名前插入的后缀（全局唯一定义）。
const ImageThumbSuffix = "_thumb"

// VideoThumbExt 视频首帧缩略图固定扩展名（与 mp4 原片不同，便于 CDN 独立缓存）。
const VideoThumbExt = "jpg"

// IsThumbObjectKey 判断 objectKey 是否已为缩略图（stem 以 ImageThumbSuffix 结尾）。
func IsThumbObjectKey(objectKey string) bool {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return false
	}
	ext := objectFileExt(key)
	if ext == "" {
		return false
	}
	stem := strings.TrimSuffix(key, "."+ext)
	return strings.HasSuffix(stem, ImageThumbSuffix)
}

// ThumbObjectKey 由原图 objectKey 派生缩略图 key：{stem}_thumb.{ext}，扩展名与原图一致。
func ThumbObjectKey(objectKey string) string {
	key := strings.TrimSpace(objectKey)
	if key == "" || IsThumbObjectKey(key) {
		return key
	}
	ext := objectFileExt(key)
	if ext == "" {
		return key
	}
	stem := strings.TrimSuffix(key, "."+ext)
	return stem + ImageThumbSuffix + "." + ext
}

// VideoThumbObjectKey 由视频原 objectKey 派生首帧缩略图 key：{stem}_thumb.jpg（扩展名固定 jpg）。
func VideoThumbObjectKey(videoObjectKey string) string {
	key := strings.TrimSpace(videoObjectKey)
	if key == "" {
		return key
	}
	ext := objectFileExt(key)
	// 已是视频 thumb key（*_thumb.jpg）时幂等返回。
	if IsThumbObjectKey(key) && ext == VideoThumbExt {
		return key
	}
	if ext == "" {
		return key
	}
	stem := strings.TrimSuffix(key, "."+ext)
	return stem + ImageThumbSuffix + "." + VideoThumbExt
}

func objectFileExt(objectKey string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(objectKey), "."))
	if ext == "jpeg" {
		return "jpg"
	}
	return ext
}
