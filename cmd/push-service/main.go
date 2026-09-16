package main

import (
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"
	"hello/internal/platform/dbcfg"
	"hello/internal/platform/loggercfg"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	pushsvc "hello/internal/services/push"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	preparePushServiceRuntime()
	ctx := gctx.New()
	if err := runtimecheck.CheckDependencies(ctx, runtimecheck.DependencyOptions{RequireRabbitMQ: false}); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	if err := pushsvc.EnsureSchema(ctx); err != nil {
		glog.Fatalf(ctx, "push schema ensure failed: %v", err)
		return
	}
	s := g.Server("push-service")
	applyPushServiceAddress(s)
	controller.RegisterPushServiceHTTP(s)
	s.Run()
}

func preparePushServiceRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.push-service.yaml")
	}
	dbcfg.ApplyGroupFromEnv("push-service", "default", "PUSH_DB_LINK", "GF_DATABASE_DEFAULT_LINK")
	rediscfg.ApplyDefaultFromEnv("push-service")
	// MUST 在 dbcfg/rediscfg（可能已初始化 logger）之后，用 GF_LOGGER_LEVEL 覆盖 yaml level。
	loggercfg.ApplyFromEnv("push-service")
}

func applyPushServiceAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("PUSH_SERVICE_ADDR"))
	if addr == "" {
		addr = ":9808"
	}
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}
