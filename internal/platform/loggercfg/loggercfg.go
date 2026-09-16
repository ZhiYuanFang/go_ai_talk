// Package loggercfg 在 yaml logger 初始化之后，用环境变量 GF_LOGGER_LEVEL 覆盖默认 logger 级别。
//
// 业务说明：GoFrame gins.Log() 只从配置文件读取 logger.level，不会用 GF_LOGGER_LEVEL 覆盖已存在的 yaml 键。
// compose 注入该变量后须显式 ApplyFromEnv，语义才与运维预期一致。
//
// 设计思路：先触达 g.Log() 完成 yaml SetConfigWithMap，再 SetLevelStr，避免被后续初始化冲掉。
//
// 使用场景：各微服务 prepare*Runtime / main 末尾（dbcfg/rediscfg 等可能已触发 logger 初始化之后）调用一次。
package loggercfg

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

const envLoggerLevel = "GF_LOGGER_LEVEL"

// ApplyFromEnv 若 GF_LOGGER_LEVEL 非空，则覆盖默认 logger 级别。
//
// Args:
//   - service: 进程名，仅用于可观测日志（如 push-service）。
//
// Side Effects: 修改默认 g.Log() 的输出级别；成功打 Warning，非法级别打 Error 且不改变级别。
func ApplyFromEnv(service string) {
	level := strings.TrimSpace(os.Getenv(envLoggerLevel))
	if level == "" {
		return
	}
	ctx := context.Background()
	service = strings.TrimSpace(service)
	if service == "" {
		service = "unknown"
	}
	// 先触达默认 logger，确保 yaml logger 段已 SetConfigWithMap。
	logger := g.Log()
	if err := logger.SetLevelStr(level); err != nil {
		g.Log().Errorf(ctx, "[loggercfg] service=%s GF_LOGGER_LEVEL=%s apply_failed err=%v (keep yaml level)",
			service, level, err)
		return
	}
	// Warning：prod 级别下仍可见，便于确认开关生效。
	g.Log().Warningf(ctx, "[loggercfg] service=%s GF_LOGGER_LEVEL=%s applied", service, level)
}
