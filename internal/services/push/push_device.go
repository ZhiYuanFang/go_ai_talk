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
//
// 业务逻辑（一部手机一个 token / 后来顶上）：
//  1. 删除库中所有相同 token 的行（含其它 wx），保证全局唯一；
//  2. 再按 (wx_id, device_key, channel) upsert 当前账号记录。
// 同用户多机：token 不同则互不影响，可并存多行。
//
// Side Effects: 可能删除他号同 token 行；写/更新本账号行。
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

	// 先清同 token（跨 wx / 跨 channel / 跨 device_key），实现后来顶上。
	if _, err := g.DB().Model(pushDeviceTable).Ctx(ctx).Where("token", token).Delete(); err != nil {
		return gerror.WrapCode(gcode.CodeDbOperationError, err, "清理冲突 token 失败")
	}

	// GoFrame：OnDuplicate 仅对 Save 生效；Insert 会忽略并变成纯 INSERT，撞 uk 即 1062。
	// OnDuplicate 的 Map/变参必须是「列名→列名」（或列名字符串），禁止传入业务值：
	// 误用 g.Map{"token": token, "updated_at": now} 会生成 VALUES(`<token正文>`) / VALUES(`unix秒`)，
	// 导致整句 Save 失败，出现「先 DELETE 成功、再 INSERT 失败」的只删不插。
	_, err := g.DB().Model(pushDeviceTable).Ctx(ctx).Data(g.Map{
		"wx_id":      wxID,
		"channel":    channel,
		"token":      token,
		"device_key": deviceKey,
		"updated_at": now,
	}).OnDuplicate("token", "updated_at").Save()
	if err != nil {
		// 常见：并发双注册撞 uk_token；也可为其它 DB 错误，客户端可重试。
		return gerror.WrapCode(gcode.CodeDbOperationError, err, "写入推送 token 失败（可能与并发注册冲突，请重试）")
	}
	return nil
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

// DeletePushDeviceByID 厂商判定 token 失效或运维手动删除后移除行。
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
	UpdatedAt int64
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
			UpdatedAt: row["updated_at"].Int64(),
		})
	}
	return out, nil
}
