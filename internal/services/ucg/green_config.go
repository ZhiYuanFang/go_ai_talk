package ucg

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// GreenConfig 阿里云 Green 内容审核配置。
type GreenConfig struct {
	Enabled         bool
	Endpoint        string
	Region          string
	AccessKeyID     string
	AccessKeySecret string
}

// LoadGreenConfig 读取 ucg.green；AK 默认可复用 OSS 凭证。
func LoadGreenConfig(ctx context.Context) GreenConfig {
	oss := LoadOSSConfig(ctx)
	cfg := GreenConfig{
		Enabled:         g.Cfg().MustGet(ctx, "ucg.green.enabled").Bool(),
		Endpoint:        strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.green.endpoint").String()),
		Region:          strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.green.region").String()),
		AccessKeyID:     strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.green.accessKeyId").String()),
		AccessKeySecret: strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.green.accessKeySecret").String()),
	}
	if v := strings.TrimSpace(os.Getenv("UCG_GREEN_ENABLED")); v == "1" || strings.EqualFold(v, "true") {
		cfg.Enabled = true
	}
	if cfg.AccessKeyID == "" {
		cfg.AccessKeyID = oss.AccessKeyID
	}
	if cfg.AccessKeySecret == "" {
		cfg.AccessKeySecret = oss.AccessKeySecret
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "green-cip.cn-beijing.aliyuncs.com"
	}
	if cfg.Region == "" {
		cfg.Region = "cn-beijing"
	}
	return cfg
}
