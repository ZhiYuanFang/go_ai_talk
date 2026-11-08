package device

import (
	"context"
	"strings"
	"unicode/utf8"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// appointmentNextRemarkMaxLen 与 API max-length:512、表 VARCHAR(512) 对齐。
const appointmentNextRemarkMaxLen = 512

// AppointmentNextView 宝宝某事件下次约定读模型（App / Internal 共用）。
type AppointmentNextView struct {
	NextAt int64
	Remark string
}

// GetAppointmentNext 读取宝宝某事件的下次约定与备注；无行视为 nextAt=0、remark=""。
// Args: deviceNo 宝宝；eventID 事件字典 ID。
// Returns: 读模型；错误仅系统失败。
// Side Effects: 读 device 库 appointment_next。
func GetAppointmentNext(ctx context.Context, deviceNo string, eventID int64) (AppointmentNextView, error) {
	var zero AppointmentNextView
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return zero, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if eventID <= 0 {
		return zero, gerror.NewCode(gcode.CodeInvalidParameter, "eventId 无效")
	}
	c := dao.AppointmentNext.Columns()
	one, err := dao.AppointmentNext.Ctx(ctx).
		Where(c.DeviceNo, deviceNo).
		Where(c.EventId, eventID).
		Fields(c.NextAt, c.Remark).
		One()
	if err != nil {
		return zero, gerror.WrapCode(gcode.CodeInternalError, err, "读取预约下次约定失败")
	}
	if one.IsEmpty() {
		return zero, nil
	}
	var row entity.AppointmentNext
	if err := one.Struct(&row); err != nil {
		return zero, gerror.WrapCode(gcode.CodeInternalError, err, "解析预约下次约定失败")
	}
	nextAt := row.NextAt
	if nextAt < 0 {
		nextAt = 0
	}
	return AppointmentNextView{
		NextAt: nextAt,
		Remark: row.Remark,
	}, nil
}

// PutAppointmentNext upsert 宝宝某事件下次约定与备注。
// 业务：nextAt>0 持久化 remark（可空串清备注）；nextAt=0 清空约定且 MUST 将 remark 置空。
// Args: deviceNo、eventID、nextAt（>=0）、remark（rune 长度 ≤512；以客户端传入为准，不 Trim）。
// Side Effects: 写 appointment_next；不触碰 predict Redis。
func PutAppointmentNext(ctx context.Context, deviceNo string, eventID, nextAt int64, remark string) error {
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
	if utf8.RuneCountInString(remark) > appointmentNextRemarkMaxLen {
		return gerror.NewCode(gcode.CodeInvalidParameter, "remark 过长")
	}
	// 无约定：备注一并清空。
	if nextAt == 0 {
		remark = ""
	}
	c := dao.AppointmentNext.Columns()
	// GoFrame：OnDuplicate 仅对 Save 生效；清空约定保留行，便于下次 upsert 覆盖。
	_, err := dao.AppointmentNext.Ctx(ctx).Data(g.Map{
		c.DeviceNo: deviceNo,
		c.EventId:  eventID,
		c.NextAt:   nextAt,
		c.Remark:   remark,
	}).OnDuplicate(c.NextAt, c.Remark).Save()
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
