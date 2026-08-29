package cash

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// NotifyUcgInviteAcquisition 通知 ucg 为码主人加获客原力 +100。
func NotifyUcgInviteAcquisition(ctx context.Context, ownerWxID int64, ref string) error {
	if ownerWxID <= 0 {
		return nil
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_SERVICE_URL")), "/")
	if base == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 UCG_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"wxId": ownerWxID,
		"ref":  strings.TrimSpace(ref),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/ucg/internal/api/force/acquire", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderInternalSecret, secret)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[cash] ucg force/acquire 调用失败 owner=%d err=%v", ownerWxID, err)
		return gerror.WrapCode(gcode.CodeInternalError, err, "ucg 获客加分失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return gerror.NewCode(gcode.CodeInternalError, "ucg 获客加分失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}
	if env.Code != 0 {
		return gerror.NewCode(gcode.CodeInternalError, "ucg 获客加分失败: "+env.Message)
	}
	return nil
}

// FetchUcgNicknames 批量拉取公开昵称（经 ucg profiles/batch）。
func FetchUcgNicknames(ctx context.Context, wxIDs []int64) (map[int64]string, error) {
	out := make(map[int64]string, len(wxIDs))
	if len(wxIDs) == 0 {
		return out, nil
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_SERVICE_URL")), "/")
	if base == "" {
		return out, gerror.NewCode(gcode.CodeInternalError, "未配置 UCG_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return out, gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	payload, _ := json.Marshal(map[string]interface{}{"wxIds": wxIDs})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/ucg/internal/api/profiles/batch", bytes.NewReader(payload))
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderInternalSecret, secret)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[cash] ucg profiles/batch 失败 err=%v", err)
		return out, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return out, gerror.NewCode(gcode.CodeInternalError, "ucg profiles/batch 失败")
	}
	var env struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				WxId     int64  `json:"wxId"`
				Nickname string `json:"nickname"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return out, err
	}
	if env.Code != 0 {
		return out, gerror.NewCode(gcode.CodeInternalError, "ucg profiles/batch 业务失败")
	}
	for _, it := range env.Data.List {
		out[it.WxId] = strings.TrimSpace(it.Nickname)
	}
	return out, nil
}
