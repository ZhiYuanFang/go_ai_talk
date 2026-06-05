package cmd

import (
	"context"

	"hello/internal/controller"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

// Main 启动 HTTP 服务；路由与处理器见 internal/controller。
var Main = gcmd.Command{
	Name:  "main",
	Usage: "main",
	Brief: "start http server",

	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		rediscfg.ApplyDefaultFromEnv("gateway")
		// 网关仍做依赖探活，避免将不可用流量放入下游链路。
		if err = runtimecheck.CheckDependencies(ctx); err != nil {
			return err
		}
		// 角色边界：gateway 仅承载入口与策略，不启动业务后台任务（由 worker 独占）。
		s := g.Server()
		controller.RegisterHTTP(s)
		s.Run()
		return nil
	},
}
