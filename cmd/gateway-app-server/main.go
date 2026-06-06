package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	"hello/internal/services/gatewayapp"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	prepareGatewayAppRuntime()
	ctx := gctx.New()
	if err := runtimecheck.CheckDependencies(ctx, runtimecheck.DependencyOptions{RequireRabbitMQ: false}); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	// 启动 Redis 订阅，将 history 变更推送至本进程 WS Hub。
	gatewayapp.StartHistoryNotifySubscriber(context.Background())

	s := g.Server("gateway-app-server")
	applyGatewayAppAddress(s)
	controller.RegisterGatewayAppHTTP(s)
	s.Run()
}

func prepareGatewayAppRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.gateway-app-server.yaml")
	}
	// MUST 在任意 g.DB("app") 之前调用；见 gatewayapp.ApplyAppDatabaseLinkFromEnv 注释。
	gatewayapp.ApplyAppDatabaseLinkFromEnv()
	rediscfg.ApplyDefaultFromEnv("gateway-app")
}

func applyGatewayAppAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("GATEWAY_APP_SERVICE_ADDR"))
	if addr == "" {
		addr = ":9702"
	}
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}
