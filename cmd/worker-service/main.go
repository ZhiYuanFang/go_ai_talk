package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"hello/internal/service"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	ctx := gctx.New()
	service.StartBackgroundWorkers(ctx)
	startWorkerHealthServer(ctx)
	select {}
}

func startWorkerHealthServer(ctx context.Context) {
	addr := strings.TrimSpace(os.Getenv("WORKER_HEALTH_ADDR"))
	if addr == "" {
		addr = ":9901"
	}
	mux := http.NewServeMux()
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
