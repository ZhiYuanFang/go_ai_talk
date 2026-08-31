package cash

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// FetchNonLeafEventCount 经 DEVICE_SERVICE_URL 拉取事件字典一级根数（历史路径/函数名含 non-leaf）。
// 语义：parent_id=0 的一级根数量，含无子根；供 catalog totalActivatableCount，与 voice「非叶子」无关。
// Side Effects: 出站 HTTP；失败由调用方决定 catalog 字段省略。
func FetchNonLeafEventCount(ctx context.Context) (int, error) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL")), "/")
	if base == "" {
		return 0, gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return 0, gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/device/internal/api/event/non-leaf-count", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set(HeaderInternalSecret, secret)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[cash] device root-count(non-leaf-count) 调用失败 err=%v", err)
		return 0, gerror.WrapCode(gcode.CodeInternalError, err, "device 一级根计数失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		glog.Warningf(ctx, "[cash] device root-count HTTP=%d body=%s", resp.StatusCode, string(body))
		return 0, gerror.NewCode(gcode.CodeInternalError, "device 一级根计数失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Count int `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return 0, gerror.WrapCode(gcode.CodeInternalError, err, "device 响应解析失败")
	}
	if env.Code != 0 {
		return 0, gerror.NewCode(gcode.CodeInternalError, "device 一级根计数失败: "+env.Message)
	}
	return env.Data.Count, nil
}

// FetchDeviceNoByWxID 经 DEVICE_SERVICE_URL 按 wx 主键取当前绑定 device_no（内部密钥）。
// 业务：邀请码兑码比对码主人与兑换者是否同一宝宝；空串表示未绑机。
// Side Effects: 出站 HTTP；失败由调用方 fail-closed。
func FetchDeviceNoByWxID(ctx context.Context, wxID int64) (string, error) {
	if wxID <= 0 {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL")), "/")
	if base == "" {
		return "", gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return "", gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	url := fmt.Sprintf("%s/device/app/api/user/internal/device-no-by-wx-id?wxId=%d", base, wxID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set(HeaderInternalSecret, secret)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[cash] device device-no-by-wx-id 调用失败 wxId=%d err=%v", wxID, err)
		return "", gerror.WrapCode(gcode.CodeInternalError, err, "查询码主人设备失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		glog.Warningf(ctx, "[cash] device device-no-by-wx-id HTTP=%d body=%s", resp.StatusCode, string(body))
		return "", gerror.NewCode(gcode.CodeInternalError, "查询码主人设备失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			DeviceNo string `json:"deviceNo"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err, "device 响应解析失败")
	}
	if env.Code != 0 {
		return "", gerror.NewCode(gcode.CodeInternalError, "查询码主人设备失败: "+env.Message)
	}
	return strings.TrimSpace(env.Data.DeviceNo), nil
}
