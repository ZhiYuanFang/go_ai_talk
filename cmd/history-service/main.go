package main

import (
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"
	"hello/internal/platform/runtimecheck"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	prepareHistoryServiceRuntime()
	ctx := gctx.New()
	// 统一 fail-fast：关键依赖不可用时 history-service 不进入监听态。
	if err := runtimecheck.CheckDependencies(ctx); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	s := g.Server("history-service")
	applyHistoryServiceAddress(s)
	controller.RegisterHistoryServiceHTTP(s)
	s.Run()
}

func prepareHistoryServiceRuntime() {
	// 默认使用 history-service 专属配置，避免误读 gateway 配置。
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.history-service.yaml")
	}
	// 允许通过环境变量覆盖数据库连接，支持独立部署时按服务拆分存储。
	if link := strings.TrimSpace(os.Getenv("HISTORY_DB_LINK")); link != "" {
		_ = os.Setenv("GF_DATABASE_DEFAULT_LINK", link)
	}
}

func applyHistoryServiceAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("HISTORY_SERVICE_ADDR"))
	if addr == "" {
		addr = ":9801"
	}
	// 兼容只传端口号的简写形式（如 9801）。
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}
