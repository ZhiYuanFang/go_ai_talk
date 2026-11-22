package history

import (
	"context"
	"fmt"
	"strings"

	deviceclient "hello/internal/clients/device"
)

// ensureHistoryDeviceRegistered 经 device 契约确认宝宝号已注册。
// 未注册（含改名后的旧号）返回可读错误，供 App/Python 写路径失败。
func ensureHistoryDeviceRegistered(ctx context.Context, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return fmt.Errorf("deviceNo 不能为空")
	}
	err := deviceclient.HTTPDeviceAdmin().EnsureRegistered(ctx, deviceNo)
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "设备未注册") || strings.Contains(msg, "未注册") {
		return fmt.Errorf("设备未注册，请确认宝宝号有效")
	}
	return fmt.Errorf("校验设备注册失败: %w", err)
}
