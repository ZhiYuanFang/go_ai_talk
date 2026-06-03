package ucg

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// OSSConfig 阿里云 OSS 直传配置（凭证可被环境变量覆盖）。
type OSSConfig struct {
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	ObjectKeyPrefix string
	CdnBaseURL      string
}

// LoadOSSConfig 读取 ucg.oss 配置；生产可用 UCG_OSS_ACCESS_KEY_* 覆盖 yaml。
func LoadOSSConfig(ctx context.Context) OSSConfig {
	cfg := OSSConfig{
		Bucket:          strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.oss.bucket").String()),
		Region:          strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.oss.region").String()),
		Endpoint:        strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.oss.endpoint").String()),
		AccessKeyID:     strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.oss.accessKeyId").String()),
		AccessKeySecret: strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.oss.accessKeySecret").String()),
		ObjectKeyPrefix: strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.oss.objectKeyPrefix").String()),
		CdnBaseURL:      strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.oss.cdnBaseUrl").String()), "/"),
	}
	if v := strings.TrimSpace(os.Getenv("UCG_OSS_ACCESS_KEY_ID")); v != "" {
		cfg.AccessKeyID = v
	}
	if v := strings.TrimSpace(os.Getenv("UCG_OSS_ACCESS_KEY_SECRET")); v != "" {
		cfg.AccessKeySecret = v
	}
	if cfg.ObjectKeyPrefix == "" {
		cfg.ObjectKeyPrefix = "social/"
	}
	if !strings.HasSuffix(cfg.ObjectKeyPrefix, "/") {
		cfg.ObjectKeyPrefix += "/"
	}
	return cfg
}
