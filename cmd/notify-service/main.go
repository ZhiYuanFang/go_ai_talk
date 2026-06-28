package main

import (
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
)

func main() {
	prepareNotifyServiceRuntime()
	s := g.Server("notify-service")
	applyNotifyServiceAddress(s)
	controller.RegisterNotifyServiceHTTP(s)
	s.Run()
}

func prepareNotifyServiceRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.notify-service.yaml")
	}
}

func applyNotifyServiceAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("NOTIFY_SERVICE_ADDR"))
	if addr == "" {
		addr = ":9806"
	}
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}
