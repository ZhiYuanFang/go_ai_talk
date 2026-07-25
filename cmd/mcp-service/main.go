package main

// cmd/mcp-service 是小智 MCP 接入服务入口。
//
// 职责：
//   - 读取环境变量（接入点 URL / token / deviceNo / 重连参数）；
//   - 校验关键配置（token / deviceNo 缺失时 fail-fast）；
//   - 构造 mcpbridge.Bridge 并阻塞运行，直到收到 SIGTERM/SIGINT。
//
// 本进程不连 MySQL、不监听 HTTP 端口、不依赖 Redis；仅通过 WebSocket 连接小智接入点，
// 并经 histsvc.DelegateTextChat 以 HTTP 委派 voice-service 完成文本对话。

import (
	"os"
	"strconv"
	"strings"
	"time"

	"hello/internal/services/mcpbridge"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

// 环境变量名定义；与 manifest/docker/.env.example 对齐。
const (
	envToken        = "XIAOZHI_MCP_TOKEN"
	envDeviceNo     = "XIAOZHI_MCP_DEVICE_NO"
	envBaseURL      = "XIAOZHI_MCP_BASE_URL"
	envReconnectMin = "XIAOZHI_MCP_RECONNECT_MIN_MS"
	envReconnectMax = "XIAOZHI_MCP_RECONNECT_MAX_MS"

	defaultBaseURL      = "wss://api.xiaozhi.me/mcp/"
	defaultReconnectMin = 2000 // 毫秒
	defaultReconnectMax = 60000
)

func main() {
	prepareRuntime()
	ctx := gctx.New()

	token := strings.TrimSpace(os.Getenv(envToken))
	deviceNo := strings.TrimSpace(os.Getenv(envDeviceNo))
	if token == "" || deviceNo == "" {
		// fail-fast：关键配置缺失时不进入重连循环，避免无意义刷日志。
		glog.Fatalf(ctx, "%s and %s are required", envToken, envDeviceNo)
		os.Exit(1)
	}

	baseURL := strings.TrimSpace(os.Getenv(envBaseURL))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	reconnectMin := parseDurationMsEnv(envReconnectMin, defaultReconnectMin)
	reconnectMax := parseDurationMsEnv(envReconnectMax, defaultReconnectMax)

	glog.Infof(ctx, "[mcp-service] starting baseURL=%s deviceNo=%s", baseURL, deviceNo)
	bridge := mcpbridge.NewBridge(baseURL, token, deviceNo, reconnectMin, reconnectMax)
	if err := bridge.Run(ctx); err != nil {
		// ctx 取消时 Run 返回 ctx.Err()，属正常退出，不输出 fatal。
		if ctx.Err() != nil {
			glog.Infof(ctx, "[mcp-service] shutdown: %v", err)
			return
		}
		glog.Fatalf(ctx, "[mcp-service] bridge run failed: %v", err)
		os.Exit(1)
	}
}

// prepareRuntime 设置默认配置文件路径，确保 GoFrame glog 等基础组件可用。
// 本进程不需要 dbcfg / rediscfg（不连库、不连 Redis）。
func prepareRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.mcp-service.yaml")
	}
}

// parseDurationMsEnv 从环境变量读取毫秒数并转为 time.Duration。
// 缺失或非法时使用 fallback，避免配置错误导致进程启动失败。
func parseDurationMsEnv(envName string, fallbackMs int) time.Duration {
	raw := strings.TrimSpace(os.Getenv(envName))
	if raw == "" {
		return time.Duration(fallbackMs) * time.Millisecond
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms <= 0 {
		// 非法值回退默认，不阻断启动。
		return time.Duration(fallbackMs) * time.Millisecond
	}
	return time.Duration(ms) * time.Millisecond
}
