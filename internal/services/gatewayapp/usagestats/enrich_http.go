package usagestats

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
)

const profileBatchChunkSize = 100

// UcgServiceBaseURL ucg-service 基址（环境变量优先）。
func UcgServiceBaseURL(ctx context.Context) string {
	u := strings.TrimSpace(os.Getenv("UCG_SERVICE_URL"))
	if u == "" {
		u = strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.ucgServiceUrl").String())
	}
	return strings.TrimRight(u, "/")
}

func gatewayInternalSecret(ctx context.Context) string {
	sec := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if sec == "" {
		sec = strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.deviceInternalSecret").String())
	}
	return sec
}

// FetchProfileNicknames 经 ucg internal batch 批量取展示昵称；失败时返回空 map（不阻断主流程）。
func FetchProfileNicknames(ctx context.Context, wxIDs []int64) map[int64]string {
	out := make(map[int64]string)
	if len(wxIDs) == 0 {
		return out
	}
	for i := 0; i < len(wxIDs); i += profileBatchChunkSize {
		end := i + profileBatchChunkSize
		if end > len(wxIDs) {
			end = len(wxIDs)
		}
		chunk := wxIDs[i:end]
		part, err := fetchProfileNicknamesChunk(ctx, chunk)
		if err != nil {
			g.Log().Warningf(ctx, "[usagestats] ucg profiles/batch 失败 err=%v", err)
			continue
		}
		for k, v := range part {
			out[k] = v
		}
	}
	return out
}

func fetchProfileNicknamesChunk(ctx context.Context, wxIDs []int64) (map[int64]string, error) {
	base := UcgServiceBaseURL(ctx)
	if base == "" {
		return nil, fmt.Errorf("UCG_SERVICE_URL 未配置")
	}
	secret := gatewayInternalSecret(ctx)
	if secret == "" {
		return nil, fmt.Errorf("DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	url := base + "/ucg/internal/api/profiles/batch"
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", secret).
		ContentJson().
		Post(ctx, url, g.Map{"wxIds": wxIDs})
	if err != nil {
		return nil, err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return nil, fmt.Errorf("ucg batch: %s", j.Get("message").String())
	}
	out := make(map[int64]string)
	for _, item := range j.Get("data.list").Array() {
		ji := gjson.New(item)
		wxID := ji.Get("wxId").Int64()
		if wxID <= 0 {
			continue
		}
		out[wxID] = strings.TrimSpace(ji.Get("nickname").String())
	}
	return out, nil
}

// DeviceWxListPage 经 device admin HTTP 拉取 wx 分页（已滤模拟用户）。
func DeviceWxListPage(ctx context.Context, page, pageSize int, q string) (list []deviceWxListItem, total, outPage, outPageSize int, err error) {
	base := gatewayapp.DeviceServiceBaseURL(ctx)
	if base == "" {
		return nil, 0, 0, 0, fmt.Errorf("DEVICE_SERVICE_URL 未配置")
	}
	pwd := gatewayapp.DeviceAdminPassword()
	if pwd == "" {
		return nil, 0, 0, 0, fmt.Errorf("DEVICE_ADMIN_PASSWORD 未配置")
	}
	reqURL := fmt.Sprintf("%s/device/admin/api/wx/list?page=%d&pageSize=%d", base, page, pageSize)
	if q = strings.TrimSpace(q); q != "" {
		reqURL += "&q=" + url.QueryEscape(q)
	}
	resp, err := gclient.New().
		SetHeader("X-Admin-Password", pwd).
		Get(ctx, reqURL)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return nil, 0, 0, 0, fmt.Errorf("device wx/list: %s", j.Get("message").String())
	}
	total = int(j.Get("data.total").Int())
	outPage = int(j.Get("data.page").Int())
	outPageSize = int(j.Get("data.pageSize").Int())
	list = make([]deviceWxListItem, 0)
	for _, row := range j.Get("data.list").Array() {
		ri := gjson.New(row)
		list = append(list, deviceWxListItem{
			Id:        ri.Get("id").Int64(),
			DeviceNo:  ri.Get("deviceNo").String(),
			Unionid:   ri.Get("unionid").String(),
			Platform:  ri.Get("platform").String(),
			Account:   ri.Get("account").String(),
			CreatedAt: ri.Get("createdAt").Int64(),
		})
	}
	return list, total, outPage, outPageSize, nil
}

// deviceWxListPageItem device wx/list 单行（gateway 编排中间结构）。
type deviceWxListItem struct {
	Id        int64
	DeviceNo  string
	Unionid   string
	Platform  string
	Account   string
	CreatedAt int64
}
