package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/model/entity"
	"hello/internal/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// HistoryCtrl 设备历史 / 建议 / 生日 API。
type HistoryCtrl struct {
	Svc   service.DeviceHistoryContract
	Voice service.VoiceContract
}

func (c *HistoryCtrl) eventRuleByID(ctx context.Context, eventID int64) (needTime bool, needQuantity bool, ok bool) {
	if eventID <= 0 {
		return false, false, false
	}
	events, err := c.Svc.ListEventOptions(ctx)
	if err != nil {
		return false, false, false
	}
	for _, ev := range events {
		if ev.Id == eventID {
			return ev.NeedTime > 0, ev.NeedQuantity > 0, true
		}
	}
	return false, false, false
}

// NewHistoryCtrl 构造 HistoryCtrl。
func NewHistoryCtrl(s service.DeviceHistoryContract, voice service.VoiceContract) *HistoryCtrl {
	return &HistoryCtrl{Svc: s, Voice: voice}
}

// List 设备历史列表。
func (c *HistoryCtrl) List(ctx context.Context, req *v1.DeviceHistoryListReq) (res *v1.DeviceHistoryListRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	items, err := c.Svc.ListHistory(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryListRes{List: items}, nil
}

// Suggest 设备建议列表。
func (c *HistoryCtrl) Suggest(ctx context.Context, req *v1.DeviceHistorySuggestReq) (res *v1.DeviceHistorySuggestRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	items, err := c.Svc.ListSuggest(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistorySuggestRes{List: items}, nil
}

// SuggestDelete 删除一条成长建议。
func (c *HistoryCtrl) SuggestDelete(ctx context.Context, req *v1.DeviceHistorySuggestDeleteReq) (res *v1.DeviceHistorySuggestDeleteRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	if err := c.Svc.DeleteSuggest(ctx, req.Id, deviceNo); err != nil {
		return nil, err
	}
	return &v1.DeviceHistorySuggestDeleteRes{}, nil
}

// EventOptions 历史事件可选项（来自数据库）。
func (c *HistoryCtrl) EventOptions(ctx context.Context, req *v1.DeviceHistoryEventOptionsReq) (res *v1.DeviceHistoryEventOptionsRes, err error) {
	_ = req
	items, err := c.Svc.ListEventOptions(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryEventOptionsRes{List: items}, nil
}

// Birthday 获取设备生日。
func (c *HistoryCtrl) Birthday(ctx context.Context, req *v1.DeviceHistoryBirthdayGetReq) (res *v1.DeviceHistoryBirthdayGetRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	birthday, sex, err := c.Svc.GetBirthday(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryBirthdayGetRes{Birthday: birthday, Sex: sex}, nil
}

// BirthdaySave 保存设备生日。
func (c *HistoryCtrl) BirthdaySave(ctx context.Context, req *v1.DeviceHistoryBirthdaySaveReq) (res *v1.DeviceHistoryBirthdaySaveRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	birthday := strings.TrimSpace(req.Birthday)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	sex := 0
	if req.Sex > 0 {
		sex = 1
	}
	if err := c.Svc.SaveBirthday(ctx, deviceNo, birthday, sex); err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryBirthdaySaveRes{}, nil
}

// Chat 文本触发智能对话（不走 STT/TTS）。
func (c *HistoryCtrl) Chat(ctx context.Context, req *v1.DeviceHistoryChatReq) (res *v1.DeviceHistoryChatRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	transcript := strings.TrimSpace(req.Transcript)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if transcript == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "transcript 不能为空")
	}
	reply, err := c.Voice.TextChat(ctx, deviceNo, transcript)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryChatRes{Reply: reply}, nil
}

// EventAdd 手动新增历史事件。
func (c *HistoryCtrl) EventAdd(ctx context.Context, req *v1.DeviceHistoryEventAddReq) (res *v1.DeviceHistoryEventAddRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	startTime := strings.TrimSpace(req.StartTime)
	endTime := strings.TrimSpace(req.EndTime)
	if needTime, needQuantity, ok := c.eventRuleByID(ctx, req.EventId); ok {
		if !needTime && !needQuantity {
			startTime = endTime
		}
	}
	id, err := c.Svc.AddHistory(ctx, entity.History{
		DeviceNo:    deviceNo,
		EventId:     req.EventId,
		EventName:   strings.TrimSpace(req.EventName),
		EventUnit:   strings.TrimSpace(req.EventUnit),
		EventNumber: int64(req.EventNumber),
		StartTime:   startTime,
		EndTime:     endTime,
		Remark:      strings.TrimSpace(req.Remark),
	})
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryEventAddRes{Id: id}, nil
}

// EventUpdate 手动修改历史事件。
func (c *HistoryCtrl) EventUpdate(ctx context.Context, req *v1.DeviceHistoryEventUpdateReq) (res *v1.DeviceHistoryEventUpdateRes, err error) {
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	startTime := strings.TrimSpace(req.StartTime)
	endTime := strings.TrimSpace(req.EndTime)
	if needTime, needQuantity, ok := c.eventRuleByID(ctx, req.EventId); ok {
		if !needTime && !needQuantity {
			startTime = endTime
		}
	}
	err = c.Svc.UpdateHistory(ctx, entity.History{
		Id:          req.Id,
		DeviceNo:    deviceNo,
		EventId:     req.EventId,
		EventName:   strings.TrimSpace(req.EventName),
		EventUnit:   strings.TrimSpace(req.EventUnit),
		EventNumber: int64(req.EventNumber),
		StartTime:   startTime,
		EndTime:     endTime,
		Remark:      strings.TrimSpace(req.Remark),
	})
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryEventUpdateRes{}, nil
}

// EventDelete 手动删除历史事件。
func (c *HistoryCtrl) EventDelete(ctx context.Context, req *v1.DeviceHistoryEventDeleteReq) (res *v1.DeviceHistoryEventDeleteRes, err error) {
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if err := c.Svc.DeleteHistory(ctx, req.Id, deviceNo); err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryEventDeleteRes{}, nil
}
