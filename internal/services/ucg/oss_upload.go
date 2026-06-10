package ucg

import (
	"context"
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// MaxMediaUploadBytes UCG 单文件上传上限（与 Flutter maxUcgVideoBytes 20MB + 余量对齐）。
const MaxMediaUploadBytes = 25 << 20

// MaxEventLogoBytes 事件 logo 单文件上限（与历史 device.eventImageMaxBytes 2MB 对齐）。
const MaxEventLogoBytes = 2 << 20

// EventLogoObjectKeyPrefix 事件 logo OSS 前缀。
const EventLogoObjectKeyPrefix = "event/"

// UploadMediaObject 服务端直传 OSS（供 Web 同域代理上传）；不写 ownership，由 register 负责。
func UploadMediaObject(ctx context.Context, wxID int64, mediaKind int, ext string, body io.Reader, size int64) (objectKey, cdnURL string, err error) {
	_ = wxID
	if size <= 0 || size > MaxMediaUploadBytes {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "文件大小无效或超过上限")
	}
	cfg := LoadOSSConfig(ctx)
	if err = validateOSSConfig(cfg); err != nil {
		return "", "", err
	}
	ext = normalizeExtension(ext)
	if ext == "" {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "extension 无效")
	}
	if mediaKind != 1 && mediaKind != 2 {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "mediaKind 无效")
	}
	objectKey = buildObjectKey(cfg.ObjectKeyPrefix, ext)
	return putOSSObject(ctx, cfg, objectKey, mediaKind, ext, body)
}

// UploadEventLogoObject 服务端直传事件 logo 至 event/ 前缀（device internal 调用）。
func UploadEventLogoObject(ctx context.Context, ext string, body io.Reader, size int64) (objectKey, cdnURL string, err error) {
	if size <= 0 || size > MaxEventLogoBytes {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "logo 文件大小无效或超过上限")
	}
	cfg := LoadOSSConfig(ctx)
	if err = validateOSSConfig(cfg); err != nil {
		return "", "", err
	}
	ext = normalizeExtension(ext)
	if ext == "" {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "extension 无效")
	}
	if !isEventLogoExt(ext) {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "logo 仅支持 png/jpg/jpeg/webp")
	}
	objectKey = buildObjectKey(EventLogoObjectKeyPrefix, ext)
	return putOSSObject(ctx, cfg, objectKey, 1, ext, body)
}

func putOSSObject(ctx context.Context, cfg OSSConfig, objectKey string, mediaKind int, ext string, body io.Reader) (string, string, error) {
	_ = ctx
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "OSS 客户端初始化失败")
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "OSS Bucket 不可用")
	}
	contentType := contentTypeForMedia(mediaKind, ext)
	if err = bucket.PutObject(objectKey, body, oss.ContentType(contentType)); err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "OSS 上传失败")
	}
	cdnURL := cfg.CdnBaseURL + "/" + strings.TrimPrefix(objectKey, "/")
	return objectKey, cdnURL, nil
}

func isEventLogoExt(ext string) bool {
	switch ext {
	case "png", "jpg", "webp":
		return true
	default:
		return false
	}
}
