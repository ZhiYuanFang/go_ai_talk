package main

import (
	"fmt"
	"hello/internal/controller"
	"hello/internal/platform/dbcfg"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	ucgsvc "hello/internal/services/ucg"
	_ "hello/internal/shared/runtime"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	prepareUcgServiceRuntime()
	ucgsvc.InitUcgPolishProfileStore()
	ctx := gctx.New()
	// 与 history/voice 一致：Redis 不可用时 fail-fast；RabbitMQ 探活失败不阻断启动（容灾）。
	if err := runtimecheck.CheckDependencies(ctx, runtimecheck.DependencyOptions{RequireRabbitMQ: true}); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	s := g.Server("ucg-service")
	applyUcgServiceAddress(s)
	controller.RegisterUcgServiceHTTP(s)
	ucgsvc.StartUcgMQConsumers(ctx)
	ucgsvc.StartAuditPublishRelayWorker(ctx) // 启动审计事件发布中继 worker，将 chat 消息等事件发布到 RabbitMQ，供审计服务消费；依赖 RabbitMQ，放在 CheckDependencies 之后启动。
	ucgsvc.StartChatPersistWorker(ctx)       // 启动聊天消息持久化 worker，从 RabbitMQ 消息队列消费聊天消息并持久化到数据库；依赖 RabbitMQ，放在 CheckDependencies 之后启动。
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
