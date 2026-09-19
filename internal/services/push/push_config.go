// Package push 系统推送域：token 注册与 APNs/HMS/MiPush 下发（宿主 push-service）。
package push

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// 推送通道常量。
const (
	PushChannelAPNs   = "apns"
	PushChannelHMS    = "hms"
	PushChannelMiPush = "mipush"
	// vivo / oppo 尚未上架：仅占位，不进入 register 白名单。
	PushChannelVivo = "vivo"
	PushChannelOppo = "oppo"
)

type pushConfig struct {
	ApnsKeyID      string
	ApnsTeamID     string
	ApnsBundleID   string
	ApnsKeyPath    string
	ApnsProduction bool

	HmsAppID     string
	HmsAppSecret string

	MipushAppID     string
	MipushAppKey    string
	MipushAppSecret string
}

func loadPushConfig(ctx context.Context) pushConfig {
	cfg := pushConfig{
		ApnsKeyID:       firstNonEmpty(os.Getenv("PUSH_APNS_KEY_ID"), cfgStr(ctx, "push.apns.keyId")),
		ApnsTeamID:      firstNonEmpty(os.Getenv("PUSH_APNS_TEAM_ID"), cfgStr(ctx, "push.apns.teamId")),
		ApnsBundleID:    firstNonEmpty(os.Getenv("PUSH_APNS_BUNDLE_ID"), cfgStr(ctx, "push.apns.bundleId")),
		ApnsKeyPath:     firstNonEmpty(os.Getenv("PUSH_APNS_KEY_PATH"), cfgStr(ctx, "push.apns.keyPath")),
		ApnsProduction:  cfgBool(ctx, "push.apns.production"),
		HmsAppID:        firstNonEmpty(os.Getenv("PUSH_HMS_APP_ID"), cfgStr(ctx, "push.hms.appId")),
		HmsAppSecret:    firstNonEmpty(os.Getenv("PUSH_HMS_APP_SECRET"), cfgStr(ctx, "push.hms.appSecret")),
		MipushAppID:     firstNonEmpty(os.Getenv("PUSH_MIPUSH_APP_ID"), cfgStr(ctx, "push.mipush.appId")),
		MipushAppKey:    firstNonEmpty(os.Getenv("PUSH_MIPUSH_APP_KEY"), cfgStr(ctx, "push.mipush.appKey")),
		MipushAppSecret: firstNonEmpty(os.Getenv("PUSH_MIPUSH_APP_SECRET"), cfgStr(ctx, "push.mipush.appSecret")),
	}
	if v := strings.TrimSpace(os.Getenv("PUSH_APNS_PRODUCTION")); v == "1" || strings.EqualFold(v, "true") {
		cfg.ApnsProduction = true
	}
	return cfg
}

func cfgStr(ctx context.Context, key string) string {
	return strings.TrimSpace(g.Cfg().MustGet(ctx, key).String())
}

func cfgBool(ctx context.Context, key string) bool {
	return g.Cfg().MustGet(ctx, key).Bool()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func apnsConfigured(cfg pushConfig) bool {
	return cfg.ApnsKeyID != "" && cfg.ApnsTeamID != "" && cfg.ApnsBundleID != "" && cfg.ApnsKeyPath != ""
}

func hmsConfigured(cfg pushConfig) bool {
	return cfg.HmsAppID != "" && cfg.HmsAppSecret != ""
}

func mipushConfigured(cfg pushConfig) bool {
	return cfg.MipushAppID != "" && cfg.MipushAppKey != "" && cfg.MipushAppSecret != ""
}
