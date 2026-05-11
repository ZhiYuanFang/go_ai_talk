package device

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
)

const (
	deviceProfileModeEnv          = "DEVICE_PROFILE_SERVICE_MODE"
	deviceProfileRemoteURLEnv     = "DEVICE_SERVICE_URL"
	deviceProfileCanaryPercentEnv = "DEVICE_PROFILE_SERVICE_CANARY_PERCENT"
	deviceProfileFailoverLocalEnv = "DEVICE_PROFILE_SERVICE_REMOTE_FAILOVER_LOCAL"
)

type DeviceProfileInfo struct {
	DeviceNo string `json:"deviceNo"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
}

type DeviceProfileContract interface {
	GetProfile(ctx context.Context, deviceNo string) (DeviceProfileInfo, error)
}

type localDeviceProfileAdapter struct{}

func (localDeviceProfileAdapter) GetProfile(ctx context.Context, deviceNo string) (DeviceProfileInfo, error) {
	var profile DeviceProfileInfo
	profile.DeviceNo = strings.TrimSpace(deviceNo)
	if cached, ok, err := deviceCache.getUserProfile(ctx, profile.DeviceNo); err == nil && ok {
		return DeviceProfileInfo{
			DeviceNo: cached.DeviceNo,
			Birthday: cached.Birthday,
			Sex:      cached.Sex,
		}, nil
	}
	var row entity.User
	err := dao.User.Ctx(ctx).
		Fields(dao.User.Columns().Birthday, dao.User.Columns().Sex).
		Where(dao.User.Columns().DeviceNo, profile.DeviceNo).
		Limit(1).
		Scan(&row)
	if err != nil {
		return DeviceProfileInfo{}, err
	}
	profile.Birthday = row.Birthday
	profile.Sex = row.Sex
	_ = deviceCache.setUserProfile(ctx, cachedUserProfile{
		DeviceNo: profile.DeviceNo,
		Birthday: profile.Birthday,
		Sex:      profile.Sex,
	})
	return profile, nil
}

type remoteDeviceProfileAdapter struct {
	baseURL string
	client  *http.Client
}

func newRemoteDeviceProfileAdapter() DeviceProfileContract {
	return &remoteDeviceProfileAdapter{
		baseURL: strings.TrimRight(strings.TrimSpace(os.Getenv(deviceProfileRemoteURLEnv)), "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (r *remoteDeviceProfileAdapter) GetProfile(ctx context.Context, deviceNo string) (DeviceProfileInfo, error) {
	if r.baseURL == "" {
		return DeviceProfileInfo{}, fmt.Errorf("device profile remote adapter not configured: missing %s", deviceProfileRemoteURLEnv)
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
	var profile DeviceProfileInfo
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, &profile); err != nil {
			return DeviceProfileInfo{}, err
		}
	}
	profile.DeviceNo = strings.TrimSpace(profile.DeviceNo)
	return profile, nil
}

type deviceProfileSwitchConfig struct {
	mode            string
	canaryPercent   int
	failoverToLocal bool
}

type deviceProfileSwitchAdapter struct {
	local  DeviceProfileContract
	remote DeviceProfileContract
	cfg    deviceProfileSwitchConfig
}

func loadDeviceProfileSwitchConfigFromEnv() deviceProfileSwitchConfig {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(deviceProfileModeEnv)))
	switch mode {
	case "remote", "canary":
	default:
		mode = "local"
	}
	canaryPercent, err := strconv.Atoi(strings.TrimSpace(os.Getenv(deviceProfileCanaryPercentEnv)))
	if err != nil {
		canaryPercent = 0
	}
	if canaryPercent < 0 {
		canaryPercent = 0
	}
	if canaryPercent > 100 {
		canaryPercent = 100
	}
	failoverRaw := strings.ToLower(strings.TrimSpace(os.Getenv(deviceProfileFailoverLocalEnv)))
	failoverToLocal := failoverRaw == "" || failoverRaw == "1" || failoverRaw == "true" || failoverRaw == "yes"
	return deviceProfileSwitchConfig{mode: mode, canaryPercent: canaryPercent, failoverToLocal: failoverToLocal}
}

func (a *deviceProfileSwitchAdapter) useRemote(deviceNo string) bool {
	switch a.cfg.mode {
	case "remote":
		return true
	case "canary":
		if a.cfg.canaryPercent <= 0 {
			return false
		}
		if a.cfg.canaryPercent >= 100 {
			return true
		}
		h := fnv.New32a()
		_, _ = h.Write([]byte(strings.TrimSpace(deviceNo)))
		return int(h.Sum32()%100) < a.cfg.canaryPercent
	default:
		return false
	}
}

func (a *deviceProfileSwitchAdapter) GetProfile(ctx context.Context, deviceNo string) (DeviceProfileInfo, error) {
	if !a.useRemote(deviceNo) {
		return a.local.GetProfile(ctx, deviceNo)
	}
	profile, err := a.remote.GetProfile(ctx, deviceNo)
	if err != nil && a.cfg.failoverToLocal {
		return a.local.GetProfile(ctx, deviceNo)
	}
	return profile, err
}

var (
	deviceProfileOnce sync.Once
	deviceProfileIns  DeviceProfileContract
)

// DeviceProfile 返回设备画像访问适配器，支持 local/remote/canary。
func DeviceProfile() DeviceProfileContract {
	deviceProfileOnce.Do(func() {
		deviceProfileIns = &deviceProfileSwitchAdapter{
			local:  localDeviceProfileAdapter{},
			remote: newRemoteDeviceProfileAdapter(),
			cfg:    loadDeviceProfileSwitchConfigFromEnv(),
		}
	})
	return deviceProfileIns
}
