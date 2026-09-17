package push

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const pushDeviceTable = "push_device"

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

// RegisterPushDevice upsert 登录用户推送 token（权威表 ai_voice_push.push_device）。
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
	// GoFrame：OnDuplicate 仅对 Save 生效；Insert 会忽略并变成纯 INSERT，撞 uk 即 1062。
	_, err := g.DB().Model(pushDeviceTable).Ctx(ctx).Data(g.Map{
		"wx_id":      wxID,
		"channel":    channel,
		"token":      token,
		"device_key": deviceKey,
		"updated_at": now,
	}).OnDuplicate(g.Map{
		"token":      token,
		"updated_at": now,
	}).Save()
	return err
}

// UnregisterPushDevice 按 wxId+deviceKey 删除（可选 channel）。
func UnregisterPushDevice(ctx context.Context, wxID int64, deviceKey, channel string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	deviceKey = strings.TrimSpace(deviceKey)
	if deviceKey == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceKey 无效")
	}
	model := g.DB().Model(pushDeviceTable).Ctx(ctx).Where("wx_id", wxID).Where("device_key", deviceKey)
	channel = strings.TrimSpace(strings.ToLower(channel))
	if channel != "" {
		if err := validatePushChannel(channel); err != nil {
			return err
		}
		model = model.Where("channel", channel)
	}
	_, err := model.Delete()
	return err
}

// DeletePushDeviceByID 厂商判定 token 失效后删除。
func DeletePushDeviceByID(ctx context.Context, id uint64) error {
	if id == 0 {
		return nil
	}
	_, err := g.DB().Model(pushDeviceTable).Ctx(ctx).Where("id", id).Delete()
	return err
}

type entityPushDevice struct {
	ID        uint64
	WxID      int64
	Channel   string
	Token     string
	DeviceKey string
}

// ListPushDevicesForWx 列出某用户全部已注册设备。
func ListPushDevicesForWx(ctx context.Context, wxID int64) ([]entityPushDevice, error) {
	if wxID <= 0 {
		return nil, nil
	}
	rows, err := g.DB().Model(pushDeviceTable).Ctx(ctx).Where("wx_id", wxID).All()
	if err != nil {
		return nil, err
	}
	out := make([]entityPushDevice, 0, len(rows))
	for _, row := range rows {
		out = append(out, entityPushDevice{
			ID:        row["id"].Uint64(),
			WxID:      row["wx_id"].Int64(),
			Channel:   row["channel"].String(),
			Token:     row["token"].String(),
			DeviceKey: row["device_key"].String(),
		})
	}
	return out, nil
}
