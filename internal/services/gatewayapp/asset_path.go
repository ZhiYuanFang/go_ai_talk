package gatewayapp

import "hello/internal/shared/assetpath"

// NormalizeAssetPath 将 download_url / logo 等字段归一为应用内路径。
func NormalizeAssetPath(raw string) string {
	return assetpath.Normalize(raw)
}
