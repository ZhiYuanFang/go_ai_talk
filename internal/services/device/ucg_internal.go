package device

import (
	"context"
	"errors"
	"strings"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
)

// UcgWxDisplay device 域向 ucg 暴露的 wx 展示字段（不含 unionid 等敏感信息）。
type UcgWxDisplay struct {
	WxId        int64  `json:"wxId"`
	Exists      bool   `json:"exists"`
	DeviceNo    string `json:"deviceNo"`
	BabyName    string `json:"babyName"`
	IpLocation  string `json:"ipLocation,omitempty"`
	IsSimulated bool   `json:"isSimulated"`
	ForceValue  int    `json:"forceValue,omitempty"`
}

var ErrUcgWxIDInvalid = errors.New("wxId 无效")

// UcgWxValidate 校验 wx 主键是否存在并返回 babyName 等展示字段。
func UcgWxValidate(ctx context.Context, wxID int64) (*UcgWxDisplay, error) {
	if wxID <= 0 {
		return nil, ErrUcgWxIDInvalid
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return nil, err
	}
	if row == nil || row.Id == 0 {
		return &UcgWxDisplay{WxId: wxID, Exists: false}, nil
	}
	babyName, _ := ucgBabyNameByDeviceNo(ctx, row.DeviceNo)
	return &UcgWxDisplay{
		WxId:        row.Id,
		Exists:      true,
		DeviceNo:    row.DeviceNo,
		BabyName:    babyName,
		IpLocation:  strings.TrimSpace(row.IpLocation),
		IsSimulated: row.IsSimulated == 1,
		ForceValue:  row.ForceValue,
	}, nil
}

// UcgWxUpdateIpLocation 更新 wx 行 IP 属地（由 ucg-service 经网关解析后写入）。
func UcgWxUpdateIpLocation(ctx context.Context, wxID int64, ipLocation string) error {
	if wxID <= 0 {
		return ErrUcgWxIDInvalid
	}
	ipLocation = strings.TrimSpace(ipLocation)
	if ipLocation == "" {
		return nil
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return err
	}
	if row == nil || row.Id == 0 {
		return ErrUcgWxIDInvalid
	}
	if strings.TrimSpace(row.IpLocation) == ipLocation {
		return nil
	}
	_, err = dao.Wx.Ctx(ctx).
		Where(dao.Wx.Columns().Id, wxID).
		Data(g.Map{dao.Wx.Columns().IpLocation: ipLocation}).
		Update()
	return err
}

// UcgWxBatch 批量查询 wx 展示字段；不存在的 wxId 返回 exists=false。
func UcgWxBatch(ctx context.Context, wxIDs []int64) ([]UcgWxDisplay, error) {
	out := make([]UcgWxDisplay, 0, len(wxIDs))
	for _, id := range wxIDs {
		item, err := UcgWxValidate(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, nil
}

// UcgWxBabyName 按 wx 主键返回 user 表 baby_name（无绑定设备时返回空字符串）。
func UcgWxBabyName(ctx context.Context, wxID int64) (string, error) {
	if wxID <= 0 {
		return "", ErrUcgWxIDInvalid
	}
	dn, err := WxDeviceNoByWxID(ctx, wxID)
	if err != nil {
		return "", err
	}
	return ucgBabyNameByDeviceNo(ctx, dn)
}

// UcgWxIncrementForceValue 作者投票成功时原力 +1。
func UcgWxIncrementForceValue(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return ErrUcgWxIDInvalid
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return err
	}
	if row == nil || row.Id == 0 {
		return ErrUcgWxIDInvalid
	}
	_, err = dao.Wx.Ctx(ctx).
		Where(dao.Wx.Columns().Id, wxID).
		Increment(dao.Wx.Columns().ForceValue, 1)
	return err
}

func ucgBabyNameByDeviceNo(ctx context.Context, deviceNo string) (string, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", nil
	}
	profile, err := DeviceProfile().GetProfile(ctx, deviceNo)
	if err != nil {
		return "", err
	}
	return profile.BabyName, nil
}
