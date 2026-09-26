package device

import (
	"context"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// GetAppointmentNextAt 读取宝宝某事件的下次约定；无行视为 0。
// Args: deviceNo 宝宝；eventID 事件字典 ID。
// Returns: nextAt（含 0）；错误仅系统失败。
// Side Effects: 读 device 库 appointment_next。
func GetAppointmentNextAt(ctx context.Context, deviceNo string, eventID int64) (int64, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if eventID <= 0 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "eventId 无效")
	}
	c := dao.AppointmentNext.Columns()
	one, err := dao.AppointmentNext.Ctx(ctx).
		Where(c.DeviceNo, deviceNo).
		Where(c.EventId, eventID).
		Fields(c.NextAt).
		One()
	if err != nil {
		return 0, gerror.WrapCode(gcode.CodeInternalError, err, "读取预约下次约定失败")
	}
	if one.IsEmpty() {
		return 0, nil
	}
	var row entity.AppointmentNext
	if err := one.Struct(&row); err != nil {
		return 0, gerror.WrapCode(gcode.CodeInternalError, err, "解析预约下次约定失败")
	}
	if row.NextAt < 0 {
		return 0, nil
	}
	return row.NextAt, nil
}

// PutAppointmentNextAt upsert 宝宝某事件下次约定；nextAt 允许 0（清空约定，保留行）。
// Args: deviceNo、eventID、nextAt（>=0）。
// Returns: 错误。
// Side Effects: 写 appointment_next；不触碰 predict Redis。
func PutAppointmentNextAt(ctx context.Context, deviceNo string, eventID, nextAt int64) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if eventID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "eventId 无效")
	}
	if nextAt < 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "nextAt 不能为负")
	}
	c := dao.AppointmentNext.Columns()
	// GoFrame：OnDuplicate 仅对 Save 生效；参数为列名（更新这些列为 VALUES）。
	// 清空约定：写入 next_at=0 并保留行，便于下次 upsert 覆盖。
	_, err := dao.AppointmentNext.Ctx(ctx).Data(g.Map{
		c.DeviceNo: deviceNo,
		c.EventId:  eventID,
		c.NextAt:   nextAt,
	}).OnDuplicate(c.NextAt).Save()
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "写入预约下次约定失败")
	}
	return nil
}

// RequireWxBoundDeviceNo 校验会话 wx 当前绑定 deviceNo 与请求一致。
// Args: wxID、请求 deviceNo。
// Returns: 不一致或不存在绑定时 NotAuthorized/业务错误。
func RequireWxBoundDeviceNo(ctx context.Context, wxID int64, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "缺少登录用户")
	}
	if deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	bound, err := WxDeviceNoByWxID(ctx, wxID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(bound) != deviceNo {
		return gerror.NewCode(gcode.CodeNotAuthorized, "deviceNo 与当前绑机不一致")
	}
	return nil
}
