package main

import (
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"
	"hello/internal/platform/dbcfg"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	voice "hello/internal/services/voice"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	prepareVoiceServiceRuntime()
	voice.InitVoiceLLMProfileStore()
	ctx := gctx.New()
	// voice-service 启动前强制检查 Redis/MQ，保证会话缓存与事件发布链路可用。
	if err := runtimecheck.CheckDependencies(ctx, runtimecheck.DependencyOptions{RequireRabbitMQ: false}); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	s := g.Server("voice-service")
	applyVoiceServiceAddress(s)
	controller.RegisterVoiceServiceHTTP(s)
	s.Run()
}

func prepareVoiceServiceRuntime() {
	// 默认使用 voice-service 专属配置，避免误读 gateway/worker 主配置。
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.voice-service.yaml")
	}
	dbcfg.ApplyGroupFromEnv("voice-service", "default", "VOICE_DB_LINK", "GF_DATABASE_DEFAULT_LINK")
	rediscfg.ApplyDefaultFromEnv("voice-service")
}

func applyVoiceServiceAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("VOICE_SERVICE_ADDR"))
	if addr == "" {
		addr = ":9802"
	}
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}

