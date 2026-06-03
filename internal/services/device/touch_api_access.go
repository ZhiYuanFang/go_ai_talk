package device

import (
	"context"
	"strings"
	"time"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
)

const touchAPIPathMaxLen = 256

// TouchLastAPIAccess 更新 user 表最近 HTTP 接口字段；设备不存在时静默（影响行数为 0）。
func (s *service) TouchLastAPIAccess(ctx context.Context, deviceNo, apiPath string, atUnixSec int64) error {
	deviceNo = strings.TrimSpace(deviceNo)
	apiPath = strings.TrimSpace(apiPath)
	if deviceNo == "" || apiPath == "" {
		return nil
	}
	if len(apiPath) > touchAPIPathMaxLen {
		apiPath = apiPath[:touchAPIPathMaxLen]
	}
	if atUnixSec <= 0 {
		atUnixSec = time.Now().Unix()
	}
	_, err := dao.User.Ctx(ctx).Where(dao.User.Columns().DeviceNo, deviceNo).Data(g.Map{
		dao.User.Columns().LastApiPath: apiPath,
		dao.User.Columns().LastApiAt:   atUnixSec,
	}).Update()
	return err
}

// TouchAPIAccessAsync 经 HTTP 异步通知 device-service internal touch；base 为空或失败时静默。
func TouchAPIAccessAsync(baseURL, deviceNo, apiPath string) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	deviceNo = strings.TrimSpace(deviceNo)
	apiPath = strings.TrimSpace(apiPath)
	if baseURL == "" || deviceNo == "" || apiPath == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		url := baseURL + "/device/internal/api/user/touch-api-access"
		_, _ = gclient.New().ContentJson().Post(ctx, url, g.Map{
			"deviceNo": deviceNo,
			"apiPath":  apiPath,
			"at":       time.Now().Unix(),
		})
	}()
}
