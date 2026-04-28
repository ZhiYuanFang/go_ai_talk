package main

import (
	"fmt"
	"os"
	"strings"

	"hello/internal/controller"

	"github.com/gogf/gf/v2/frame/g"
)

func main() {
	prepareHistoryServiceRuntime()
	s := g.Server("history-service")
	applyHistoryServiceAddress(s)
	controller.RegisterHistoryServiceHTTP(s)
	s.Run()
}

func prepareHistoryServiceRuntime() {
	// Force history-service to use its dedicated config file by default.
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.history-service.yaml")
	}
	// Allow independent DB ownership via runtime override without touching monolith config.
	if link := strings.TrimSpace(os.Getenv("HISTORY_DB_LINK")); link != "" {
		_ = os.Setenv("GF_DATABASE_DEFAULT_LINK", link)
	}
}

func applyHistoryServiceAddress(s interface{ SetAddr(address string) }) {
	addr := strings.TrimSpace(os.Getenv("HISTORY_SERVICE_ADDR"))
	if addr == "" {
		addr = ":9801"
	}
	// support shorthand value like 9801
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	s.SetAddr(addr)
}
