// 设备画像 HTTP 读客户端（remote-only，供 voice 等跨进程使用）。
package device

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// DeviceProfileInfo 设备画像快照。
type DeviceProfileInfo struct {
	DeviceNo string `json:"deviceNo"`
	BabyName string `json:"babyName"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
}

// DeviceProfileContract 画像读取契约。
type DeviceProfileContract interface {
	GetProfile(ctx context.Context, deviceNo string) (DeviceProfileInfo, error)
}

type remoteDeviceProfileAdapter struct {
	baseURL string
	client  *http.Client
}

var (
	profileOnce sync.Once
	profileIns  DeviceProfileContract
)

// HTTPDeviceProfile 返回仅经 DEVICE_SERVICE_URL 拉取画像的客户端。
func HTTPDeviceProfile() DeviceProfileContract {
	profileOnce.Do(func() {
		profileIns = &remoteDeviceProfileAdapter{
			baseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL")), "/"),
			client:  &http.Client{Timeout: 5 * time.Second},
		}
	})
	return profileIns
}

func (r *remoteDeviceProfileAdapter) GetProfile(ctx context.Context, deviceNo string) (DeviceProfileInfo, error) {
	if r.baseURL == "" {
		return DeviceProfileInfo{}, fmt.Errorf("device profile remote adapter not configured: missing DEVICE_SERVICE_URL")
	}
	u, err := url.Parse(r.baseURL + "/device/app/api/user/get")
	if err != nil {
		return DeviceProfileInfo{}, err
	}
	q := u.Query()
	q.Set("deviceNo", strings.TrimSpace(deviceNo))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return DeviceProfileInfo{}, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return DeviceProfileInfo{}, err
	}
	defer resp.Body.Close()
	var env struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return DeviceProfileInfo{}, err
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		if strings.TrimSpace(env.Message) == "" {
			return DeviceProfileInfo{}, fmt.Errorf("device profile remote call failed: status=%d", resp.StatusCode)
		}
		return DeviceProfileInfo{}, fmt.Errorf("device profile remote call failed: %s", strings.TrimSpace(env.Message))
	}
	var info DeviceProfileInfo
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, &info); err != nil {
			return DeviceProfileInfo{}, err
		}
	}
	info.DeviceNo = strings.TrimSpace(info.DeviceNo)
	return info, nil
}
