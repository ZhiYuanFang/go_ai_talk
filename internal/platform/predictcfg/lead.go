// Package predictcfg 预测临近推送相关跨进程配置读取。
//
// 业务说明：leadSeconds 同时服务于 voice 普通 predict_imminent 叫醒（fireAt=nextAt-lead）
// 与 history 统一 Live Activity 预测段「进窗」过滤；MUST 同 env、同默认，禁止各服务漂移。
package predictcfg

import (
	"os"
	"strconv"
	"strings"
)

const (
	// LeadSecondsEnv 与 voice 历史常量 VOICE_PREDICT_IMMINENT_LEAD_SECONDS 对齐。
	LeadSecondsEnv = "VOICE_PREDICT_IMMINENT_LEAD_SECONDS"
	// LeadSecondsDefault 默认提前 5 分钟（秒）。
	LeadSecondsDefault int64 = 300
)

// LeadSeconds 读取提前叫醒 / LA 进窗秒数。
//
// 业务：与 MQ fireAt = nextAt - lead、普通推送去重 TTL 同语义；非法或未配置回退 300。
// Args: 无（读环境变量 LeadSecondsEnv）
// Returns: 非负整数秒；0 表示立即到点（火点已过或配置为 0）。
func LeadSeconds() int64 {
	v := strings.TrimSpace(os.Getenv(LeadSecondsEnv))
	if v == "" {
		return LeadSecondsDefault
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n < 0 {
		return LeadSecondsDefault
	}
	return n
}
