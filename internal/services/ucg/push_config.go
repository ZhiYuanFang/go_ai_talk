package ucg

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
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
		ApnsKeyID:        firstNonEmpty(os.Getenv("UCG_APNS_KEY_ID"), cfgStr(ctx, "ucg.push.apns.keyId")),
		ApnsTeamID:       firstNonEmpty(os.Getenv("UCG_APNS_TEAM_ID"), cfgStr(ctx, "ucg.push.apns.teamId")),
		ApnsBundleID:     firstNonEmpty(os.Getenv("UCG_APNS_BUNDLE_ID"), cfgStr(ctx, "ucg.push.apns.bundleId")),
		ApnsKeyPath:      firstNonEmpty(os.Getenv("UCG_APNS_KEY_PATH"), cfgStr(ctx, "ucg.push.apns.keyPath")),
		ApnsProduction:   cfgBool(ctx, "ucg.push.apns.production"),
		HmsAppID:         firstNonEmpty(os.Getenv("UCG_HMS_APP_ID"), cfgStr(ctx, "ucg.push.hms.appId")),
		HmsAppSecret:     firstNonEmpty(os.Getenv("UCG_HMS_APP_SECRET"), cfgStr(ctx, "ucg.push.hms.appSecret")),
		MipushAppID:      firstNonEmpty(os.Getenv("UCG_MIPUSH_APP_ID"), cfgStr(ctx, "ucg.push.mipush.appId")),
		MipushAppKey:     firstNonEmpty(os.Getenv("UCG_MIPUSH_APP_KEY"), cfgStr(ctx, "ucg.push.mipush.appKey")),
		MipushAppSecret:  firstNonEmpty(os.Getenv("UCG_MIPUSH_APP_SECRET"), cfgStr(ctx, "ucg.push.mipush.appSecret")),
	}
	if v := strings.TrimSpace(os.Getenv("UCG_APNS_PRODUCTION")); v == "1" || strings.EqualFold(v, "true") {
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
