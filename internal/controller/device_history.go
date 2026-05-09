package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/model/entity"
	contracts "hello/internal/services/contracts"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// HistoryCtrl 设备历史 / 建议 / 生日 API。
type HistoryCtrl struct {
	Svc   contracts.DeviceHistoryContract
	Voice contracts.VoiceContract
}

func (c *HistoryCtrl) eventRuleByID(ctx context.Context, eventID int64) (needQuantity bool, ok bool) {
	if eventID <= 0 {
		return false, false
	}
	events, err := c.Svc.ListEventOptions(ctx)
	if err != nil {
		return false, false
	}
	for _, ev := range events {
		if ev.Id == eventID {
			return ev.NeedQuantity > 0, true
		}
	}
	return false, false
}

// NewHistoryCtrl 构造 HistoryCtrl。
func NewHistoryCtrl(s contracts.DeviceHistoryContract, voice contracts.VoiceContract) *HistoryCtrl {
	return &HistoryCtrl{Svc: s, Voice: voice}
}

// List 设备历史列表。
func (c *HistoryCtrl) List(ctx context.Context, req *v1.DeviceHistoryListReq) (res *v1.DeviceHistoryListRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	result, err := c.Svc.ListHistoryPage(ctx, deviceNo, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryListRes{List: result.List, Total: result.Total, Page: result.Page, PageSize: result.PageSize}, nil
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
	id, err := c.Svc.AddHistory(ctx, entity.History{
		DeviceNo:    deviceNo,
		EventId:     req.EventId,
		EventName:   strings.TrimSpace(req.EventName),
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
	err = c.Svc.UpdateHistory(ctx, entity.History{
		Id:          req.Id,
		DeviceNo:    deviceNo,
		EventId:     req.EventId,
		EventName:   strings.TrimSpace(req.EventName),
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

// EventLatest 查询最近一条历史事件（内部契约）。
func (c *HistoryCtrl) EventLatest(ctx context.Context, req *v1.DeviceHistoryLatestReq) (res *v1.DeviceHistoryLatestRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	item, err := c.Svc.GetLatestHistory(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryLatestRes{Item: item}, nil
}

// EventEndLatest 条件结束最近一条历史事件（内部契约）。
func (c *HistoryCtrl) EventEndLatest(ctx context.Context, req *v1.DeviceHistoryEndLatestReq) (res *v1.DeviceHistoryEndLatestRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if req.EventId <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "eventId 无效")
	}
	updated, err := c.Svc.EndLatestHistoryIfMatch(ctx, deviceNo, req.EventId, strings.TrimSpace(req.EndTime))
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryEndLatestRes{Updated: updated}, nil
}
