package v2

import "embed"

// APIMetaFS 嵌入 api/v2 下 g.Meta 源文件，供 gateway-app apiregistry 加载 v2 路由 summary。
//
//go:embed *.go
var APIMetaFS embed.FS
