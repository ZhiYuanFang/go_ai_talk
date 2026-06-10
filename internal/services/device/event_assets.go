package device

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"hello/internal/shared/eventlogo"
)

const defaultEventLogoMaxBytes = 2 << 20 // 2MB

var eventColorPattern = regexp.MustCompile(`(?i)^#([0-9a-f]{3}|[0-9a-f]{6})$`)

// EventImageMaxBytes 单张 logo 大小上限。
func EventImageMaxBytes(ctx context.Context) int64 {
	_ = ctx
	return defaultEventLogoMaxBytes
}

// ValidateEventColor 校验色值；空串表示不设置颜色。
func ValidateEventColor(color string) error {
	color = strings.TrimSpace(color)
	if color == "" {
		return nil
	}
	if !eventColorPattern.MatchString(color) {
		return fmt.Errorf("color 须为 #RGB 或 #RRGGBB 格式")
	}
	return nil
}

// UploadEventLogo 经 ucg-service 上传 logo 至 OSS，返回 objectKey（event/ 前缀）。
func UploadEventLogo(ctx context.Context, originalFilename string, src io.Reader, size int64) (string, error) {
	maxBytes := EventImageMaxBytes(ctx)
	ext, err := eventLogoExt(originalFilename)
	if err != nil {
		return "", err
	}
	lr := io.LimitReader(src, maxBytes+1)
	data, err := io.ReadAll(lr)
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxBytes {
		return "", fmt.Errorf("logo 文件过大，上限 %d 字节", maxBytes)
	}
	if size > 0 && size > maxBytes {
		return "", fmt.Errorf("logo 文件过大，上限 %d 字节", maxBytes)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("logo 文件为空")
	}
	name := strings.TrimSpace(originalFilename)
	if name == "" {
		name = "logo." + ext
	}
	objectKey, err := ucgUpload().uploadEventLogoViaUcg(ctx, name, data)
	if err != nil {
		return "", err
	}
	return eventlogo.NormalizeObjectKey(objectKey), nil
}

func eventLogoExt(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		if ext == ".jpeg" {
			return "jpg", nil
		}
		return strings.TrimPrefix(ext, "."), nil
	default:
		return "", fmt.Errorf("logo 仅支持 png/jpg/jpeg/webp")
	}
}

// NormalizeEventLogoStored 归一化写入 DB 的 objectKey。
func NormalizeEventLogoStored(logoPath string) string {
	return eventlogo.NormalizeObjectKey(logoPath)
}
