package ucg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/grand"
)

const presignExpireSeconds = 900

// PresignUpload 生成 social/ 前缀 objectKey 与 PUT 预签名 URL，并记录 ucg_media_upload 所有权。
func PresignUpload(ctx context.Context, wxID int64, mediaKind int, ext string) (uploadURL, objectKey, cdnURL string, headers map[string]string, err error) {
	cfg := LoadOSSConfig(ctx)
	if err = validateOSSConfig(cfg); err != nil {
		return "", "", "", nil, err
	}
	ext = normalizeExtension(ext)
	if ext == "" {
		return "", "", "", nil, gerror.NewCode(gcode.CodeInvalidParameter, "extension 无效")
	}
	objectKey = buildObjectKey(cfg.ObjectKeyPrefix, ext)
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return "", "", "", nil, gerror.WrapCode(gcode.CodeInternalError, err, "OSS 客户端初始化失败")
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return "", "", "", nil, gerror.WrapCode(gcode.CodeInternalError, err, "OSS Bucket 不可用")
	}
	contentType := contentTypeForMedia(mediaKind, ext)
	opts := []oss.Option{
		oss.ContentType(contentType),
	}
	signedURL, err := bucket.SignURL(objectKey, oss.HTTPPut, presignExpireSeconds, opts...)
	if err != nil {
		return "", "", "", nil, gerror.WrapCode(gcode.CodeInternalError, err, "生成预签名 URL 失败")
	}
	cdnURL = cfg.CdnBaseURL + "/" + strings.TrimPrefix(objectKey, "/")
	headers = map[string]string{
		"Content-Type": contentType,
	}
	if logErr := LogMediaUpload(ctx, wxID, objectKey, mediaKind); logErr != nil {
		return "", "", "", nil, gerror.WrapCode(gcode.CodeInternalError, logErr, "记录媒体所有权失败")
	}
	return signedURL, objectKey, cdnURL, headers, nil
}

func validateOSSConfig(cfg OSSConfig) error {
	if cfg.Bucket == "" || cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return gerror.NewCode(gcode.CodeInternalError, "OSS 配置不完整")
	}
	return nil
}

func normalizeExtension(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	ext = strings.TrimPrefix(ext, ".")
	switch ext {
	case "jpeg":
		return "jpg"
	case "mpeg":
		return "mp4"
	default:
		return ext
	}
}

func buildObjectKey(prefix, ext string) string {
	now := time.Now()
	return fmt.Sprintf("%s%04d/%02d/%s.%s", prefix, now.Year(), int(now.Month()), grand.S(32), ext)
}

func contentTypeForMedia(mediaKind int, ext string) string {
	if mediaKind == 2 {
		switch ext {
		case "mp4":
			return "video/mp4"
		case "mov":
			return "video/quicktime"
		default:
			return "video/mp4"
		}
	}
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}
