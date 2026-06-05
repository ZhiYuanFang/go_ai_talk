package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"hello/internal/platform/dbcfg"
	"hello/internal/platform/rediscfg"
	"hello/internal/platform/runtimecheck"
	_ "hello/internal/shared/runtime"
	async "hello/internal/services/async"
	"hello/internal/services/workeroutbox"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	prepareWorkerServiceRuntime()
	ctx := gctx.New()
	// worker 同样依赖 Redis/MQ，启动前即校验，避免任务处理中途大面积失败。
	if err := runtimecheck.CheckDependencies(ctx); err != nil {
		glog.Fatalf(ctx, "dependency check failed: %v", err)
		return
	}
	startBackgroundWorkers(ctx)
	// 对外暴露最小健康探针，便于容器编排系统判断进程存活。
	startWorkerHealthServer(ctx)
	select {}
}

func prepareWorkerServiceRuntime() {
	// 默认使用 worker-service 专属配置，避免误读 gateway 主配置。
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.worker-service.yaml")
	}
	dbcfg.ApplyGroupFromEnv("worker-service", "outbox", "WORKER_OUTBOX_DB_LINK", "GF_DATABASE_OUTBOX_LINK")
	rediscfg.ApplyDefaultFromEnv("worker-service")
}

var workerOnce sync.Once

func startBackgroundWorkers(ctx context.Context) {
	workerOnce.Do(func() {
		async.StartVoiceTaskConsumer(ctx)
		async.StartDomainOutboxRelay(ctx)
	})
}

func startWorkerHealthServer(ctx context.Context) {
	addr := strings.TrimSpace(os.Getenv("WORKER_HEALTH_ADDR"))
	if addr == "" {
		addr = ":9901"
	}
	mux := http.NewServeMux()
	workeroutbox.RegisterOutboxEnqueueHandler(mux)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			glog.Warningf(ctx, "worker health server listen failed: %v", err)
		}
	}()
}
