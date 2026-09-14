package ucg

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// OSSConfig 阿里云 OSS 配置（凭证 / endpoint 可被环境变量覆盖）。
// 业务说明：默认 caidoukeji + 杭州内网；客户端一律经服务端 multipart 上传，不再签发 App 预签名直传。
type OSSConfig struct {
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	ObjectKeyPrefix string
	CdnBaseURL      string
}

// LoadOSSConfig 读取 ucg.oss 配置。
//
// 业务逻辑：
//   - AccessKey 仅允许 env（UCG_OSS_ACCESS_KEY_*）注入，yaml 明文留空；
//   - Endpoint 可用 UCG_OSS_ENDPOINT 覆盖（本地非 VPC 联调公网杭州 endpoint）；
//   - objectKeyPrefix 缺省 social/。
//
// Args: ctx — 配置读取上下文。
// Returns: 合并后的 OSSConfig。
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
	// 生产凭证 MUST 经 env；禁止依赖 yaml 明文。
	if v := strings.TrimSpace(os.Getenv("UCG_OSS_ACCESS_KEY_ID")); v != "" {
		cfg.AccessKeyID = v
	}
	if v := strings.TrimSpace(os.Getenv("UCG_OSS_ACCESS_KEY_SECRET")); v != "" {
		cfg.AccessKeySecret = v
	}
	// 非 VPC 本地联调：覆盖为公网杭州 endpoint（如 oss-cn-hangzhou.aliyuncs.com）。
	if v := strings.TrimSpace(os.Getenv("UCG_OSS_ENDPOINT")); v != "" {
		cfg.Endpoint = v
	}
	if cfg.ObjectKeyPrefix == "" {
		cfg.ObjectKeyPrefix = "social/"
	}
	if !strings.HasSuffix(cfg.ObjectKeyPrefix, "/") {
		cfg.ObjectKeyPrefix += "/"
	}
	return cfg
}
