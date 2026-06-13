package ucg

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// GreenConfig 阿里云 Green 内容审核配置；进程启动时 LoadGreenConfig 读一次，Green() 单例缓存。
type GreenConfig struct {
	Enabled         bool   // false → noopGreenModerator，不调 API、全 pass（止血开关）
	Endpoint        string // Green OpenAPI endpoint，默认 green-cip.cn-beijing.aliyuncs.com
	Region          string
	AccessKeyID     string // 空则复用 OSS AK
	AccessKeySecret string
}

// LoadGreenConfig 从 manifest/config.ucg-service.yaml 的 ucg.green 段读取配置。
// 注意：改 yaml 后需 rebuild/restart ucg-service 才生效（Docker COPY 配置，非 volume）。
func LoadGreenConfig(ctx context.Context) GreenConfig {
	oss := LoadOSSConfig(ctx) // AK 兜底来源
	cfg := GreenConfig{
		Enabled:         g.Cfg().MustGet(ctx, "ucg.green.enabled").Bool(), // 生产 true；测试可设 false 止血
		Endpoint:        strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.green.endpoint").String()),
		Region:          strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.green.region").String()),
		AccessKeyID:     strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.green.accessKeyId").String()),
		AccessKeySecret: strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.green.accessKeySecret").String()),
	}
	// 环境变量仅能强制开启，不能通过 env 关闭 yaml 里的 enabled=true
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
