package device

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
)

const maxDeviceModelLen = 128

// ErrWxDeviceModelInvalid 型号文案非法（过长或含控制字符）。
var ErrWxDeviceModelInvalid = errors.New("deviceModel 非法")

// WxUpdateDeviceModel 更新当前账号设备型号展示文案。
//
// 业务逻辑：客户端原文入库，不做品牌枚举；空串清空；与库中相同则跳过 UPDATE。
// Args: wxID 须为正整数；deviceModel 最长 128（trim 后字节长度）。
// Returns: 参数/账号错误；成功时 nil。
// Side Effects: 可能 UPDATE wx.device_model。
func WxUpdateDeviceModel(ctx context.Context, wxID int64, deviceModel string) error {
	if wxID <= 0 {
		return ErrWxDeactivateWxIDInvalid
	}
	deviceModel = strings.TrimSpace(deviceModel)
	if len(deviceModel) > maxDeviceModelLen {
		return ErrWxDeviceModelInvalid
	}
	// 禁止控制字符污染运维展示列。
	for _, r := range deviceModel {
		if unicode.IsControl(r) {
			return ErrWxDeviceModelInvalid
		}
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return err
	}
	if row == nil || row.Id == 0 {
		return ErrWxDeactivateNotFound
	}
	if strings.TrimSpace(row.DeviceModel) == deviceModel {
		return nil
	}
	_, err = dao.Wx.Ctx(ctx).
		Where(dao.Wx.Columns().Id, wxID).
		Data(g.Map{dao.Wx.Columns().DeviceModel: deviceModel}).
		Update()
	return err
}
