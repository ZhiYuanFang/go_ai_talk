package ucg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/grand"
)

// PresignUpload 已停用：禁止向客户端签发 OSS 预签名上传 URL。
//
// 业务说明：桶迁至 caidoukeji 后服务端使用杭州内网 endpoint；若再签发 SignURL，
// 手机会拿到不可达的内网地址。全客户端改走 POST /media/upload 服务端 PutObject。
//
// Args: ctx/wxID/mediaKind/ext — 保留原签名以兼容 controller 调用，参数均忽略。
// Returns: 明确业务错误，且 uploadURL 恒为空。
// Side Effects: 无（不初始化 OSS 客户端、不 SignURL）。
func PresignUpload(ctx context.Context, wxID int64, mediaKind int, ext string) (uploadURL, objectKey, cdnURL string, headers map[string]string, err error) {
	_ = ctx
	_ = wxID
	_ = mediaKind
	_ = ext
	return "", "", "", nil, gerror.NewCode(
		gcode.CodeNotSupported,
		"预签名直传已停用，请使用 POST /ucg/app/api/media/upload（经网关 multipart）",
	)
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
