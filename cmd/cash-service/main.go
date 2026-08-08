package main

import (
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"
	"hello/internal/platform/dbcfg"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	"hello/internal/services/cash"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	prepareCashServiceRuntime()
	ctx := gctx.New()
	// cash-service 启动期 fail-fast：依赖不可用则不对外提供支付/权益。
	if err := runtimecheck.CheckDependencies(ctx, runtimecheck.DependencyOptions{RequireRabbitMQ: false}); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	if err := cash.EnsureSchema(ctx); err != nil {
		glog.Fatalf(ctx, "cash schema ensure failed: %v", err)
		return
	}
	s := g.Server("cash-service")
	applyCashServiceAddress(s)
	controller.RegisterCashServiceHTTP(s)
	s.Run()
}

func prepareCashServiceRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.cash-service.yaml")
	}
	dbcfg.ApplyGroupFromEnv("cash-service", "default", "CASH_DB_LINK", "GF_DATABASE_DEFAULT_LINK")
	rediscfg.ApplyDefaultFromEnv("cash-service")
}

func applyCashServiceAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("CASH_SERVICE_ADDR"))
	if addr == "" {
		// 9806 已由 notify-service 占用，现金域使用 9807。
		addr = ":9807"
	}
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}
