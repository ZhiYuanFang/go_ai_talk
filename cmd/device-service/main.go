package main

import (
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"
	device "hello/internal/services/device"
	"hello/internal/platform/dbcfg"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	prepareDeviceServiceRuntime()
	ctx := gctx.New()
	// device-service 与其他领域服务保持一致的 fail-fast 启动语义。
	if err := runtimecheck.CheckDependencies(ctx, runtimecheck.DependencyOptions{RequireRabbitMQ: false}); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	if err := device.EnsureWxIsSimulatedColumn(ctx); err != nil {
		glog.Fatalf(ctx, "wx schema ensure failed: %v", err)
		return
	}
	s := g.Server("device-service")
	applyDeviceServiceAddress(s)
	controller.RegisterDeviceServiceHTTP(s)
	s.Run()
}

func prepareDeviceServiceRuntime() {
	// 默认使用 device-service 专属配置，确保服务边界清晰。
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.device-service.yaml")
	}
	dbcfg.ApplyGroupFromEnv("device-service", "default", "DEVICE_DB_LINK", "GF_DATABASE_DEFAULT_LINK")
	rediscfg.ApplyDefaultFromEnv("device-service")
}

func applyDeviceServiceAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("DEVICE_SERVICE_ADDR"))
	if addr == "" {
		addr = ":9803"
	}
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}

