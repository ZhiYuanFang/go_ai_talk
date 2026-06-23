package ucg

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"hello/internal/shared/mediacdn"
)

const thumbResizeProcessBase = "image/auto-orient,1/resize,m_lfit,w_200/quality,q_90"

// EnsureImageThumb 确保图片原图 objectKey 在 OSS 上存在对应物理缩略图对象（幂等）。
// 经 OSS 图片处理拉取字节后 PUT 至 ThumbObjectKey；失败时返回错误供 register 阻断。
func EnsureImageThumb(ctx context.Context, objectKey string) error {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "objectKey 无效")
	}
	if mediacdn.IsThumbObjectKey(key) {
		return nil
	}
	ext := imageExtFromObjectKey(key)
	if !isImageObjectExt(ext) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "非图片 objectKey")
	}

	thumbKey := mediacdn.ThumbObjectKey(key)
	bucket, _, err := openOSSBucket(ctx)
	if err != nil {
		return err
	}

	thumbExists, err := bucket.IsObjectExist(thumbKey)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "OSS HEAD 缩略图失败")
	}
	if thumbExists {
		return nil
	}

	origExists, err := bucket.IsObjectExist(key)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "OSS HEAD 原图失败")
	}
	if !origExists {
		return gerror.NewCode(gcode.CodeInvalidParameter, "OSS 原图不存在")
	}

	process := thumbProcessForExt(ext)
	body, err := bucket.GetObject(key, oss.Process(process))
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "OSS 拉取缩略图字节失败")
	}
	defer func() { _ = body.Close() }()

	data, err := io.ReadAll(body)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "读取缩略图字节失败")
	}
	if len(data) == 0 {
		return gerror.NewCode(gcode.CodeInternalError, "缩略图字节为空")
	}

	contentType := contentTypeForMedia(1, ext)
	if err = bucket.PutObject(thumbKey, bytes.NewReader(data), oss.ContentType(contentType)); err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "OSS 上传缩略图失败")
	}
	return nil
}

// IsImageObjectKey 判断是否为受支持的图片原图 objectKey（非 thumb、非视频）。
func IsImageObjectKey(objectKey string) bool {
	key := strings.TrimSpace(objectKey)
	if key == "" || mediacdn.IsThumbObjectKey(key) {
		return false
	}
	return isImageObjectExt(imageExtFromObjectKey(key))
}

// deletePairedThumbObject 删除原图对应的缩略图对象；对象不存在时忽略。
func deletePairedThumbObject(bucket *oss.Bucket, objectKey string) error {
	thumbKey := mediacdn.ThumbObjectKey(objectKey)
	if thumbKey == "" || thumbKey == objectKey {
		return nil
	}
	if err := bucket.DeleteObject(thumbKey); err != nil {
		if ossErr, ok := err.(oss.ServiceError); ok && ossErr.StatusCode == 404 {
			return nil
		}
		return gerror.WrapCode(gcode.CodeInternalError, err, "OSS 删除缩略图失败")
	}
	return nil
}

func openOSSBucket(ctx context.Context) (*oss.Bucket, OSSConfig, error) {
	cfg := LoadOSSConfig(ctx)
	if err := validateOSSConfig(cfg); err != nil {
		return nil, cfg, err
	}
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, cfg, gerror.WrapCode(gcode.CodeInternalError, err, "OSS 客户端初始化失败")
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, cfg, gerror.WrapCode(gcode.CodeInternalError, err, "OSS Bucket 不可用")
	}
	return bucket, cfg, nil
}

func thumbProcessForExt(ext string) string {
	switch ext {
	case "png":
		return thumbResizeProcessBase + "/format,png"
	case "webp":
		return thumbResizeProcessBase + "/format,webp"
	case "gif":
		return thumbResizeProcessBase + "/format,gif"
	default:
		return thumbResizeProcessBase
	}
}

func imageExtFromObjectKey(objectKey string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(objectKey), "."))
	if ext == "jpeg" {
		return "jpg"
	}
	return ext
}

func isImageObjectExt(ext string) bool {
	switch ext {
	case "jpg", "png", "webp", "gif":
		return true
	default:
		return false
	}
}
