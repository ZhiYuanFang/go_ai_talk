package cash

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	ucgclient "hello/internal/clients/ucg"
	"hello/internal/shared/featurelogo"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

const defaultFeatureLogoMaxBytes = 2 << 20 // 2MB，与事件/ucg 上限对齐

var featureColorPattern = regexp.MustCompile(`(?i)^#([0-9a-f]{3}|[0-9a-f]{6})$`)

// ValidateFeatureColor 校验功能主色；空串允许（防御路径，种子后通常非空）。
func ValidateFeatureColor(color string) error {
	color = strings.TrimSpace(color)
	if color == "" {
		return nil
	}
	if !featureColorPattern.MatchString(color) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "color 须为 #RGB 或 #RRGGBB 格式")
	}
	return nil
}

// UploadFeatureLogo 经 clients/ucg 上传功能 logo，返回归一化 objectKey。
//
// Args: originalFilename 用于扩展名校验；src/size 为文件内容。
// Returns: feature/ 前缀 objectKey。
func UploadFeatureLogo(ctx context.Context, originalFilename string, src io.Reader, size int64) (string, error) {
	ext, err := featureLogoExt(originalFilename)
	if err != nil {
		return "", err
	}
	lr := io.LimitReader(src, defaultFeatureLogoMaxBytes+1)
	data, err := io.ReadAll(lr)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInvalidParameter, err, "读取 logo 失败")
	}
	if int64(len(data)) > defaultFeatureLogoMaxBytes {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, fmt.Sprintf("logo 文件过大，上限 %d 字节", defaultFeatureLogoMaxBytes))
	}
	if size > 0 && size > defaultFeatureLogoMaxBytes {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, fmt.Sprintf("logo 文件过大，上限 %d 字节", defaultFeatureLogoMaxBytes))
	}
	if len(data) == 0 {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "logo 文件为空")
	}
	name := strings.TrimSpace(originalFilename)
	if name == "" {
		name = "logo." + ext
	}
	objectKey, _, err := ucgclient.UploadFeatureLogo(ctx, name, data)
	if err != nil {
		return "", err
	}
	return featurelogo.NormalizeObjectKey(objectKey), nil
}

// FeatureLogoCdnURL 将 objectKey 映射为 CDN URL（Admin 上传响应）。
func FeatureLogoCdnURL(ctx context.Context, objectKey string) string {
	return featurelogo.CdnURL(ctx, objectKey)
}

func featureLogoExt(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		if ext == ".jpeg" {
			return "jpg", nil
		}
		return strings.TrimPrefix(ext, "."), nil
	default:
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "logo 仅支持 png/jpg/jpeg/webp")
	}
}
