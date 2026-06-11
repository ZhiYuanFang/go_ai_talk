package v1

import "embed"

// APIMetaFS 嵌入 api/v1 下 g.Meta 源文件，供 gateway-app apiregistry 编译期加载（Docker 镜像无需 COPY api/）。
//
//go:embed *.go
var APIMetaFS embed.FS
