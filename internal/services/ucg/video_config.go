package ucg

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// VideoTransformV1 Web Phase 1 宽验真直传管线版本（非 canonical）。
const VideoTransformV1 = "v1"

// VideoTransformV2 原生/sim 服务端 canonical 管线版本。
const VideoTransformV2 = "v2"

// VideoConfig 视频验真与转码运行时配置。
type VideoConfig struct {
	MaxTranscodeConcurrency int
	TranscodeTimeoutSec     int
}

// LoadVideoConfig 读取 ucg.video.* 配置。
func LoadVideoConfig(ctx context.Context) VideoConfig {
	cfg := VideoConfig{
		MaxTranscodeConcurrency: g.Cfg().MustGet(ctx, "ucg.video.maxTranscodeConcurrency").Int(),
		TranscodeTimeoutSec:     g.Cfg().MustGet(ctx, "ucg.video.transcodeTimeoutSec").Int(),
	}
	if cfg.MaxTranscodeConcurrency <= 0 {
		cfg.MaxTranscodeConcurrency = 2
	}
	if cfg.TranscodeTimeoutSec <= 0 {
		cfg.TranscodeTimeoutSec = 120
	}
	return cfg
}

// allowedVideoTransformVersion 判断视频 register 允许的 transformVersion。
func allowedVideoTransformVersion(version string) bool {
	return version == VideoTransformV1 || version == VideoTransformV2
}
