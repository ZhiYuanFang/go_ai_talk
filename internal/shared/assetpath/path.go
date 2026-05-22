package assetpath

import (
	"net/url"
	"strings"
)

// Normalize 将库内或历史绝对 URL 归一为应用内路径（以 / 开头，不含 scheme）。
func Normalize(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		if u, err := url.Parse(raw); err == nil {
			p := strings.TrimSpace(u.Path)
			if p != "" {
				return Normalize(p)
			}
		}
	}
	if !strings.HasPrefix(raw, "/") {
		return "/" + raw
	}
	return raw
}
