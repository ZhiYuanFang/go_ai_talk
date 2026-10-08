package main

// cmd/xiaozhi-mcp-service 是小智 MCP 多设备接入服务入口。
//
// 职责：
//   - 内部 HTTP：Upsert/Remove/health（供 device-service 写路径推送）；
//   - Manager：按绑定集维护 N 条出站 Bridge（replicas=1）；
//   - 启动全量拉取 + 可选低频 reconcile；
//   - 可选 env 单 token fallback（DB 为空时便于灰度）。
//
// 本进程不连 MySQL；经 DEVICE_SERVICE_URL 拉绑定；经 VOICE_* 走 /voice/chat/ws。

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"hello/internal/platform/loggercfg"
	"hello/internal/services/mcpbridge"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	envBaseURL         = "XIAOZHI_MCP_BASE_URL"
	envReconnectMin    = "XIAOZHI_MCP_RECONNECT_MIN_MS"
	envReconnectMax    = "XIAOZHI_MCP_RECONNECT_MAX_MS"
	envInternalAddr    = "XIAOZHI_MCP_INTERNAL_ADDR"
	envReconcileMs     = "XIAOZHI_MCP_RECONCILE_INTERVAL_MS"
	envFallbackToken   = "XIAOZHI_MCP_TOKEN"
	envFallbackDevice  = "XIAOZHI_MCP_DEVICE_NO"
	defaultBaseURL     = "wss://api.xiaozhi.me/mcp/"
	defaultReconnectMin = 2000
	defaultReconnectMax = 60000
	defaultInternalAddr = ":9809"
	defaultReconcileMs  = 600000 // 10 分钟
)

func main() {
	prepareRuntime()
	ctx, stop := signal.NotifyContext(gctx.New(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	baseURL := strings.TrimSpace(os.Getenv(envBaseURL))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	reconnectMin := parseDurationMsEnv(envReconnectMin, defaultReconnectMin)
	reconnectMax := parseDurationMsEnv(envReconnectMax, defaultReconnectMax)
	internalAddr := strings.TrimSpace(os.Getenv(envInternalAddr))
	if internalAddr == "" {
		internalAddr = defaultInternalAddr
	}
	reconcileEvery := parseDurationMsEnvAllowZero(envReconcileMs, defaultReconcileMs)

	mgr := mcpbridge.NewManager(baseURL, reconnectMin, reconnectMax)
	glog.Infof(ctx, "[xiaozhi-mcp-service] starting baseURL=%s internalAddr=%s reconcile=%v",
		baseURL, internalAddr, reconcileEvery)

	// 启动全量对齐（失败不退出：依赖后续 reconcile / 写路径推送）。
	if err := mcpbridge.PullAndReconcile(ctx, mgr); err != nil {
		glog.Warningf(ctx, "[xiaozhi-mcp-service] initial reconcile failed err=%v", err)
		tryEnvFallback(ctx, mgr)
	} else if mgr.ActiveCount() == 0 {
		tryEnvFallback(ctx, mgr)
	}

	go mcpbridge.StartReconcileLoop(ctx, mgr, reconcileEvery)

	httpSrv := &mcpbridge.InternalHTTP{Manager: mgr}
	go func() {
		if err := mcpbridge.RunInternalHTTP(ctx, internalAddr, httpSrv); err != nil && ctx.Err() == nil {
			glog.Errorf(ctx, "[xiaozhi-mcp-service] internal HTTP exited err=%v", err)
		}
	}()

	<-ctx.Done()
	glog.Infof(context.Background(), "[xiaozhi-mcp-service] shutdown")
}

// tryEnvFallback 当 DB 绑定为空或拉取失败时，可选注入单条 env Bridge（迁移期）。
func tryEnvFallback(ctx context.Context, mgr *mcpbridge.Manager) {
	token := strings.TrimSpace(os.Getenv(envFallbackToken))
	deviceNo := strings.TrimSpace(os.Getenv(envFallbackDevice))
	if token == "" || deviceNo == "" {
		glog.Infof(ctx, "[xiaozhi-mcp-service] no DB bindings and no env fallback; waiting for Upsert")
		return
	}
	glog.Infof(ctx, "[xiaozhi-mcp-service] using env fallback deviceNo=%s", deviceNo)
	mgr.Upsert(ctx, mcpbridge.BindingSpec{Id: 0, McpToken: token, DeviceNo: deviceNo})
}

func prepareRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.xiaozhi-mcp-service.yaml")
	}
	loggercfg.ApplyFromEnv("xiaozhi-mcp-service")
}

func parseDurationMsEnv(envName string, fallbackMs int) time.Duration {
	raw := strings.TrimSpace(os.Getenv(envName))
	if raw == "" {
		return time.Duration(fallbackMs) * time.Millisecond
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms <= 0 {
		return time.Duration(fallbackMs) * time.Millisecond
	}
	return time.Duration(ms) * time.Millisecond
}

// parseDurationMsEnvAllowZero 允许 0 表示关闭周期 reconcile。
func parseDurationMsEnvAllowZero(envName string, fallbackMs int) time.Duration {
	raw := strings.TrimSpace(os.Getenv(envName))
	if raw == "" {
		return time.Duration(fallbackMs) * time.Millisecond
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms < 0 {
		return time.Duration(fallbackMs) * time.Millisecond
	}
	return time.Duration(ms) * time.Millisecond
}
