package device

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	defaultEventImageStorageDir = "/ai_talk_images/"
	defaultEventImageMaxBytes   = 2 << 20 // 2MB
	eventLogoURLPrefix          = "/ai_talk_images/"
)

var eventColorPattern = regexp.MustCompile(`(?i)^#([0-9a-f]{3}|[0-9a-f]{6})$`)

// EventImageStorageDir 事件 logo 落盘目录，默认 /ai_talk_images/。
func EventImageStorageDir(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("DEVICE_EVENT_IMAGE_STORAGE_DIR")); v != "" {
		return ensureDirSuffix(v)
	}
	s := strings.TrimSpace(g.Cfg().MustGet(ctx, "device.eventImageStorageDir").String())
	if s == "" {
		return defaultEventImageStorageDir
	}
	return ensureDirSuffix(s)
}

// EventImageMaxBytes 单张 logo 大小上限。
func EventImageMaxBytes(ctx context.Context) int64 {
	if v := strings.TrimSpace(os.Getenv("DEVICE_EVENT_IMAGE_MAX_BYTES")); v != "" {
		if n, err := parsePositiveInt64(v); err == nil {
			return n
		}
	}
	n := g.Cfg().MustGet(ctx, "device.eventImageMaxBytes").Int64()
	if n <= 0 {
		return defaultEventImageMaxBytes
	}
	return n
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

// SaveEventLogo 将上传图片写入存储目录，返回库内路径如 /ai_talk_images/event_1_xxx.png。
func SaveEventLogo(ctx context.Context, eventID int64, originalFilename string, src io.Reader, size int64) (string, error) {
	if eventID <= 0 {
		return "", fmt.Errorf("事件ID无效")
	}
	ext, err := eventLogoExt(originalFilename)
	if err != nil {
		return "", err
	}
	dir := filepath.Clean(EventImageStorageDir(ctx))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建 logo 目录失败: %w", err)
	}
	serverName := fmt.Sprintf("event_%d_%d%s", eventID, time.Now().UnixNano(), ext)
	if !eventLogoFilenameSafe(serverName) {
		return "", fmt.Errorf("非法文件名")
	}
	dest := filepath.Join(dir, serverName)
	out, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("保存 logo 失败: %w", err)
	}
	defer func() { _ = out.Close() }()
	maxBytes := EventImageMaxBytes(ctx)
	lr := io.LimitReader(src, maxBytes+1)
	written, err := io.Copy(out, lr)
	if err != nil {
		_ = os.Remove(dest)
		return "", err
	}
	if written > maxBytes {
		_ = os.Remove(dest)
		return "", fmt.Errorf("logo 文件过大，上限 %d 字节", maxBytes)
	}
	if size > 0 && size > maxBytes {
		_ = os.Remove(dest)
		return "", fmt.Errorf("logo 文件过大，上限 %d 字节", maxBytes)
	}
	return eventLogoURLPrefix + serverName, nil
}

// EventLogoAbsPath 由库内路径解析磁盘绝对路径；路径非法时返回错误。
func EventLogoAbsPath(ctx context.Context, logoPath string) (string, error) {
	logoPath = strings.TrimSpace(logoPath)
	if logoPath == "" || !strings.HasPrefix(logoPath, eventLogoURLPrefix) {
		return "", fmt.Errorf("logo 路径无效")
	}
	name := strings.TrimPrefix(logoPath, eventLogoURLPrefix)
	if name == "" || !eventLogoFilenameSafe(name) {
		return "", fmt.Errorf("logo 文件名非法")
	}
	dir := filepath.Clean(EventImageStorageDir(ctx))
	abs := filepath.Join(dir, name)
	rel, err := filepath.Rel(dir, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("logo 路径越界")
	}
	return abs, nil
}

// EventLogoContentType 按扩展名返回 Content-Type。
func EventLogoContentType(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func eventLogoExt(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		return ext, nil
	default:
		return "", fmt.Errorf("logo 仅支持 png/jpg/jpeg/webp")
	}
}

func eventLogoFilenameSafe(name string) bool {
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return false
	}
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-' {
			continue
		}
		return false
	}
	return true
}

func ensureDirSuffix(dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return defaultEventImageStorageDir
	}
	if !strings.HasSuffix(dir, "/") && !strings.HasSuffix(dir, "\\") {
		dir += string(os.PathSeparator)
	}
	return dir
}

func parsePositiveInt64(v string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid")
	}
	return n, nil
}
