package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/model/entity"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// VerifyAdminPassword 内部校验管理口令（供 voice HTTP 客户端）。
func (c *DeviceInternalCtrl) VerifyAdminPassword(ctx context.Context, req *v1.DeviceInternalVerifyAdminPasswordReq) (res *v1.DeviceInternalVerifyAdminPasswordRes, err error) {
	_ = c
	ok := device.DeviceAdmin().VerifyPassword(req.Password)
	return &v1.DeviceInternalVerifyAdminPasswordRes{OK: ok}, nil
}

// RegisterDevice 内部注册设备（免 X-Admin-Password，依赖网络隔离）。
func (c *DeviceInternalCtrl) RegisterDevice(ctx context.Context, req *v1.DeviceInternalRegisterReq) (res *v1.DeviceInternalRegisterRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	activeTime, err := device.DeviceAdmin().Register(ctx, deviceNo)
	if err != nil {
		if err == device.ErrDeviceExists {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceInternalRegisterRes{DeviceNo: deviceNo, ActiveTime: activeTime}, nil
}

// UserEnsure 确保设备已注册。
func (c *DeviceInternalCtrl) UserEnsure(ctx context.Context, req *v1.DeviceInternalUserEnsureReq) (res *v1.DeviceInternalUserEnsureRes, err error) {
	_ = c
	if err := device.DeviceAdmin().EnsureRegistered(ctx, strings.TrimSpace(req.DeviceNo)); err != nil {
		if err == device.ErrDeviceNotRegistered {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceInternalUserEnsureRes{}, nil
}

// UserLastTalk 更新最近对话摘要。
func (c *DeviceInternalCtrl) UserLastTalk(ctx context.Context, req *v1.DeviceInternalUserLastTalkReq) (res *v1.DeviceInternalUserLastTalkRes, err error) {
	_ = c
	if err := device.DeviceAdmin().UpdateLastTalk(ctx, strings.TrimSpace(req.DeviceNo), req.Ask, req.Answer); err != nil {
		return nil, err
	}
	return &v1.DeviceInternalUserLastTalkRes{}, nil
}

// UserList 设备列表。
func (c *DeviceInternalCtrl) UserList(ctx context.Context, req *v1.DeviceInternalUserListReq) (res *v1.DeviceInternalUserListRes, err error) {
	_ = c
	_ = req
	items, err := device.DeviceAdmin().List(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalUserListRes{List: items}, nil
}

// EventAddInternal 新增事件字典。
func (c *DeviceInternalCtrl) EventAddInternal(ctx context.Context, req *v1.DeviceInternalEventAddReq) (res *v1.DeviceInternalEventAddRes, err error) {
	_ = c
	if _, err := device.DeviceAdmin().AddEvent(ctx, strings.TrimSpace(req.Name), req.EventType, req.ExtraNames, "", "", req.ParentId); err != nil {
		switch err {
		case device.ErrEventExists:
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		case device.ErrEventNotFound:
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		default:
			return nil, err
		}
	}
	return &v1.DeviceInternalEventAddRes{}, nil
}

// EventUpdateInternal 更新事件。
func (c *DeviceInternalCtrl) EventUpdateInternal(ctx context.Context, req *v1.DeviceInternalEventUpdateReq) (res *v1.DeviceInternalEventUpdateRes, err error) {
	_ = c
	if err := device.DeviceAdmin().UpdateEvent(ctx, req.Id, strings.TrimSpace(req.Name), req.EventType, req.ExtraNames, "", ""); err != nil {
		switch err {
		case device.ErrEventExists:
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		case device.ErrEventNotFound:
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		default:
			return nil, err
		}
	}
	return &v1.DeviceInternalEventUpdateRes{}, nil
}

// EventDeleteInternal 删除事件。
func (c *DeviceInternalCtrl) EventDeleteInternal(ctx context.Context, req *v1.DeviceInternalEventDeleteReq) (res *v1.DeviceInternalEventDeleteRes, err error) {
	_ = c
	if err := device.DeviceAdmin().DeleteEvent(ctx, req.Id); err != nil {
		switch err {
		case device.ErrEventNotFound:
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		case device.ErrEventHasChildren:
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		default:
			return nil, err
		}
	}
	return &v1.DeviceInternalEventDeleteRes{}, nil
}

// QAList 内部 QA 列表。
func (c *DeviceInternalCtrl) QAList(ctx context.Context, req *v1.DeviceInternalQAListReq) (res *v1.DeviceInternalQAListRes, err error) {
	_ = c
	_ = req
	result, err := device.DeviceAdmin().ListQAPage(ctx, 1, 100)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalQAListRes{List: result.List}, nil
}

// ActionList 内部动作列表。
func (c *DeviceInternalCtrl) ActionList(ctx context.Context, req *v1.DeviceInternalActionListReq) (res *v1.DeviceInternalActionListRes, err error) {
	_ = c
	_ = req
	items, err := device.DeviceAdmin().ListActionsForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalActionListRes{List: items}, nil
}

// ActionUpdateInternal 更新动作。
func (c *DeviceInternalCtrl) ActionUpdateInternal(ctx context.Context, req *v1.DeviceInternalActionUpdateReq) (res *v1.DeviceInternalActionUpdateRes, err error) {
	_ = c
	if err := device.DeviceAdmin().UpdateAction(ctx, req.Id, strings.TrimSpace(req.Name), strings.TrimSpace(req.TargetType)); err != nil {
		if err == device.ErrActionNotFound {
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceInternalActionUpdateRes{}, nil
}

// ActionDeleteInternal 删除动作。
func (c *DeviceInternalCtrl) ActionDeleteInternal(ctx context.Context, req *v1.DeviceInternalActionDeleteReq) (res *v1.DeviceInternalActionDeleteRes, err error) {
	_ = c
	if err := device.DeviceAdmin().DeleteAction(ctx, req.Id); err != nil {
		if err == device.ErrActionNotFound {
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceInternalActionDeleteRes{}, nil
}

// VoiceInsertAction 语音写入动作。
func (c *DeviceInternalCtrl) VoiceInsertAction(ctx context.Context, req *v1.DeviceInternalVoiceActionReq) (res *v1.DeviceInternalVoiceActionRes, err error) {
	_ = c
	if err := device.DeviceAdmin().InsertVoiceActionRecord(ctx, req.Name, req.TargetType); err != nil {
		return nil, err
	}
	return &v1.DeviceInternalVoiceActionRes{}, nil
}

// VoiceEventNeedle 语音按名插入/回读事件。
func (c *DeviceInternalCtrl) VoiceEventNeedle(ctx context.Context, req *v1.DeviceInternalVoiceEventNeedleReq) (res *v1.DeviceInternalVoiceEventNeedleRes, err error) {
	_ = c
	item, err := device.DeviceAdmin().InsertOrGetEventByNeedle(ctx, req.Needle, req.EventType)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalVoiceEventNeedleRes{Item: item}, nil
}

// VoiceEventDeepSeek DeepSeek 抽取落库。
func (c *DeviceInternalCtrl) VoiceEventDeepSeek(ctx context.Context, req *v1.DeviceInternalVoiceEventDeepSeekReq) (res *v1.DeviceInternalVoiceEventDeepSeekRes, err error) {
	_ = c
	out := entity.Event{
		Name:       strings.TrimSpace(req.Name),
		ExtraNames: strings.TrimSpace(req.ExtraNames),
		EventType:  device.NormalizeEventType(req.EventType),
	}
	item, target, err := device.DeviceAdmin().ApplyDeepSeekEventExtractPersistence(ctx, out)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalVoiceEventDeepSeekRes{Item: item, TargetName: target}, nil
}
