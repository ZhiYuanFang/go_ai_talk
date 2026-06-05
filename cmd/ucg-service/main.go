package main

import (
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"
	"hello/internal/platform/dbcfg"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	ucgsvc "hello/internal/services/ucg"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	prepareUcgServiceRuntime()
	ctx := gctx.New()
	// 与 history/voice 一致：Redis、RabbitMQ 不可用时 fail-fast，避免半启动态。
	if err := runtimecheck.CheckDependencies(ctx); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	s := g.Server("ucg-service")
	applyUcgServiceAddress(s)
	controller.RegisterUcgServiceHTTP(s)
	ucgsvc.StartAuditWorker(ctx)
	ucgsvc.StartRecommendWorker(ctx)
	s.Run()
}

func prepareUcgServiceRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.ucg-service.yaml")
	}
	dbcfg.ApplyGroupFromEnv("ucg-service", "default", "UCG_DB_LINK", "GF_DATABASE_DEFAULT_LINK")
	rediscfg.ApplyDefaultFromEnv("ucg-service")
}

func applyUcgServiceAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("UCG_SERVICE_ADDR"))
	if addr == "" {
		addr = ":9804"
	}
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}
