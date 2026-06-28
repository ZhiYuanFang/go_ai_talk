package ucg

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// UploadMediaResult 服务端直传 OSS 结果。
type UploadMediaResult struct {
	ObjectKey         string
	CdnURL            string
	ContentHash       string // 视频：OSS 最终对象字节 SHA-256 hex（直传为原始字节，转码为 v2 字节）
	TransformVersion  string // 视频：v1 直传 | v2 服务端转码，供 register 配对 contentHash
}

// UploadMediaObject 服务端直传 OSS（供 Web 同域代理上传）；不写 ownership，由 register 负责。
func UploadMediaObject(ctx context.Context, wxID int64, mediaKind int, ext string, body io.Reader, size int64) (*UploadMediaResult, error) {
	_ = wxID
	if size <= 0 || size > MaxMediaUploadBytes {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "文件大小无效或超过上限")
	}
	cfg := LoadOSSConfig(ctx)
	if err := validateOSSConfig(cfg); err != nil {
		return nil, err
	}
	ext = normalizeExtension(ext)
	if ext == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "extension 无效")
	}
	if mediaKind != 1 && mediaKind != 2 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "mediaKind 无效")
	}

	data, err := readAllLimited(body, MaxMediaUploadBytes)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "读取上传文件失败")
	}
	if mediaKind == 2 {
		// Web 视频三分支：A v1 直传 | B v1 失败且可解码→转码 v2 | C 不可解码 4xx
		if err := ValidateVideoBytes(VideoTransformV1, data); err != nil {
			if probeErr := ProbeVideoDecodable(data); probeErr != nil {
				return nil, probeErr
			}
			result, transErr := UploadVideoTranscodedObject(ctx, data)
			if transErr != nil {
				return nil, transErr
			}
			result.TransformVersion = VideoTransformV2
			return result, nil
		}
		ext = "mp4"
	}

	objectKey := buildObjectKey(cfg.ObjectKeyPrefix, ext)
	objectKey, cdnURL, err := putOSSObjectBytes(ctx, cfg, objectKey, mediaKind, ext, data)
	if err != nil {
		return nil, err
	}
	res := &UploadMediaResult{ObjectKey: objectKey, CdnURL: cdnURL}
	if mediaKind == 2 {
		res.ContentHash = sha256HexBytes(data)
		res.TransformVersion = VideoTransformV1
	}
	return res, nil
}

// UploadVideoTranscodedObject internal 转码上传：NormalizeVideo → PUT mp4（v2 canonical）。
func UploadVideoTranscodedObject(ctx context.Context, input []byte) (*UploadMediaResult, error) {
	if len(input) == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "视频为空")
	}
	out, err := NormalizeVideo(ctx, input)
	if err != nil {
		return nil, err
	}
	cfg := LoadOSSConfig(ctx)
	if err = validateOSSConfig(cfg); err != nil {
		return nil, err
	}
	objectKey := buildObjectKey(cfg.ObjectKeyPrefix, "mp4")
	objectKey, cdnURL, err := putOSSObjectBytes(ctx, cfg, objectKey, 2, "mp4", out)
	if err != nil {
		return nil, err
	}
	return &UploadMediaResult{
		ObjectKey:        objectKey,
		CdnURL:           cdnURL,
		ContentHash:      sha256HexBytes(out),
		TransformVersion: VideoTransformV2,
	}, nil
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
	data, err := readAllLimited(body, MaxEventLogoBytes)
	if err != nil {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "读取 logo 失败")
	}
	objectKey = buildObjectKey(EventLogoObjectKeyPrefix, ext)
	return putOSSObjectBytes(ctx, cfg, objectKey, 1, ext, data)
}

func putOSSObjectBytes(ctx context.Context, cfg OSSConfig, objectKey string, mediaKind int, ext string, data []byte) (string, string, error) {
	return putOSSObject(ctx, cfg, objectKey, mediaKind, ext, bytes.NewReader(data))
}

func putOSSObject(ctx context.Context, cfg OSSConfig, objectKey string, mediaKind int, ext string, body io.Reader) (string, string, error) {
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
	if mediaKind == 1 {
		if thumbErr := EnsureImageThumb(ctx, objectKey); thumbErr != nil {
			return "", "", thumbErr
		}
	} else if mediaKind == 2 {
		if thumbErr := EnsureVideoThumb(ctx, objectKey); thumbErr != nil {
			return "", "", thumbErr
		}
	}
	cdnURL := cfg.CdnBaseURL + "/" + strings.TrimPrefix(objectKey, "/")
	return objectKey, cdnURL, nil
}

func sha256HexBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func isEventLogoExt(ext string) bool {
	switch ext {
	case "png", "jpg", "webp":
		return true
	default:
		return false
	}
}
