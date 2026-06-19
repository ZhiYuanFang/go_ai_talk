package main

import (
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"
	"hello/internal/platform/dbcfg"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	simuser "hello/internal/services/simuser"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	prepareSimUserServiceRuntime()
	ctx := gctx.New()
	if err := runtimecheck.CheckDependencies(ctx, runtimecheck.DependencyOptions{RequireRabbitMQ: false}); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	simuser.InitAIModel()
	if err := simuser.EnsureSchema(ctx); err != nil {
		glog.Fatalf(ctx, "sim schema ensure failed: %v", err)
		return
	}
	flags := simuser.LoadRuntimeFlags(ctx)
	simuser.StartScheduler(ctx, flags)

	s := g.Server("sim-user-service")
	applySimUserServiceAddress(s)
	controller.RegisterSimUserServiceHTTP(s)
	s.Run()
}

func prepareSimUserServiceRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.sim-user-service.yaml")
	}
	dbcfg.ApplyGroupFromEnv("sim-user-service", "default", "SIM_DB_LINK", "GF_DATABASE_DEFAULT_LINK")
	rediscfg.ApplyDefaultFromEnv("sim-user-service")
}

func applySimUserServiceAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("SIM_USER_SERVICE_ADDR"))
	if addr == "" {
		addr = ":9805"
	}
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}
