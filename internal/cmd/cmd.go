package cmd

import (
	"context"

	"hello/internal/controller"
	"hello/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

// Main 启动 HTTP 服务；路由与处理器见 internal/controller。
var Main = gcmd.Command{
	Name:  "main",
	Usage: "main",
	Brief: "start http server",

	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		prepareDomainDBRuntime()
		service.StartBackgroundWorkers(ctx)
		s := g.Server()
		controller.RegisterHTTP(s)
		s.Run()
		return nil
	},
}
