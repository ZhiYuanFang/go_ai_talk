package ucg

import (
	"context"
	"strings"
	"time"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	PushChannelAPNs    = "apns"
	PushChannelHMS     = "hms"
	PushChannelMiPush  = "mipush"
)

var validPushChannels = map[string]struct{}{
	PushChannelAPNs:   {},
	PushChannelHMS:    {},
	PushChannelMiPush: {},
}

func validatePushChannel(channel string) error {
	channel = strings.TrimSpace(strings.ToLower(channel))
	if _, ok := validPushChannels[channel]; !ok {
		return gerror.NewCode(gcode.CodeInvalidParameter, "channel 必须是 apns、hms 或 mipush")
	}
	return nil
}

// RegisterPushDevice upserts a push token for the authenticated wxId.
func RegisterPushDevice(ctx context.Context, wxID int64, channel, token, deviceKey string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	channel = strings.TrimSpace(strings.ToLower(channel))
	if err := validatePushChannel(channel); err != nil {
		return err
	}
	token = strings.TrimSpace(token)
	deviceKey = strings.TrimSpace(deviceKey)
	if token == "" || len(token) > 512 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "token 无效")
	}
	if deviceKey == "" || len(deviceKey) > 64 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceKey 无效")
	}
	now := time.Now().Unix()
	cols := dao.UcgPushDevice.Columns()
	model := dao.UcgPushDevice.Ctx(ctx).
		Where(cols.WxId, wxID).
		Where(cols.DeviceKey, deviceKey).
		Where(cols.Channel, channel)
	cnt, err := model.Count()
	if err != nil {
		return err
	}
	data := g.Map{
		cols.WxId:      wxID,
		cols.Channel:   channel,
		cols.Token:     token,
		cols.DeviceKey: deviceKey,
		cols.UpdatedAt: now,
	}
	if cnt > 0 {
		_, err = model.Data(g.Map{
			cols.Token:     token,
			cols.UpdatedAt: now,
		}).Update()
		return err
	}
	_, err = dao.UcgPushDevice.Ctx(ctx).Data(data).Insert()
	return err
}

// UnregisterPushDevice removes push rows for wxId + deviceKey (optional channel filter).
func UnregisterPushDevice(ctx context.Context, wxID int64, deviceKey, channel string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	deviceKey = strings.TrimSpace(deviceKey)
	if deviceKey == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceKey 无效")
	}
	model := dao.UcgPushDevice.Ctx(ctx).
		Where(dao.UcgPushDevice.Columns().WxId, wxID).
		Where(dao.UcgPushDevice.Columns().DeviceKey, deviceKey)
	channel = strings.TrimSpace(strings.ToLower(channel))
	if channel != "" {
		if err := validatePushChannel(channel); err != nil {
			return err
		}
		model = model.Where(dao.UcgPushDevice.Columns().Channel, channel)
	}
	_, err := model.Delete()
	return err
}

// DeletePushDeviceByID removes a stale device row after vendor invalid-token response.
func DeletePushDeviceByID(ctx context.Context, id uint64) error {
	if id == 0 {
		return nil
	}
	_, err := dao.UcgPushDevice.Ctx(ctx).Where(dao.UcgPushDevice.Columns().Id, id).Delete()
	return err
}

// ListPushDevicesForWx returns all registered devices for fan-out.
func ListPushDevicesForWx(ctx context.Context, wxID int64) ([]entityPushDevice, error) {
	if wxID <= 0 {
		return nil, nil
	}
	rows, err := dao.UcgPushDevice.Ctx(ctx).
		Where(dao.UcgPushDevice.Columns().WxId, wxID).
		All()
	if err != nil {
		return nil, err
	}
	out := make([]entityPushDevice, 0, len(rows))
	for _, row := range rows {
		out = append(out, entityPushDevice{
			ID:        row[dao.UcgPushDevice.Columns().Id].Uint64(),
			WxID:      row[dao.UcgPushDevice.Columns().WxId].Int64(),
			Channel:   row[dao.UcgPushDevice.Columns().Channel].String(),
			Token:     row[dao.UcgPushDevice.Columns().Token].String(),
			DeviceKey: row[dao.UcgPushDevice.Columns().DeviceKey].String(),
		})
	}
	return out, nil
}

type entityPushDevice struct {
	ID        uint64
	WxID      int64
	Channel   string
	Token     string
	DeviceKey string
}
