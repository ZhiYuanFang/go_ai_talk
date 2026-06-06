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

// UploadMediaObject 服务端直传 OSS（供 Web 等同域代理上传，规避浏览器对 OSS 的 CORS 预检）。
func UploadMediaObject(ctx context.Context, mediaKind int, ext string, body io.Reader, size int64) (objectKey, cdnURL string, err error) {
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
	cdnURL = cfg.CdnBaseURL + "/" + strings.TrimPrefix(objectKey, "/")
	return objectKey, cdnURL, nil
}
