package controller

import (
	"context"
	"errors"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/model/entity"
	contracts "hello/internal/services/contracts"
	device "hello/internal/services/device"
	"hello/internal/services/gatewayapp"
	histsvc "hello/internal/services/history"
	voice "hello/internal/services/voice"
	"hello/internal/shared/eventlogo"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// mapVoiceAIQuotaErr 将 voice 额度/登录错误转为 HTTP 业务码，供 history chat 等 HTTP 路径返回 40301/40302。
func mapVoiceAIQuotaErr(err error) error {
	if err == nil {
		return nil
	}
	var qErr *voice.VoiceAIQuotaError
	if errors.As(err, &qErr) && qErr != nil {
		switch qErr.Code {
		case device.CodeAINotLoggedIn:
			return gerror.NewCode(device.GCodeAINotLoggedIn(), qErr.Message)
		case device.CodeAIQuotaExhausted:
			return gerror.NewCode(device.GCodeAIQuotaExhausted(), qErr.Message)
		}
	}
	return err
}

// HistoryCtrl 设备历史 / 建议 / 生日 API。
type HistoryCtrl struct {
	Svc   contracts.DeviceHistoryContract
	Voice contracts.VoiceContract
}

// canonicalEventNameForRow 有 eventId 时与事件主档 name 对齐写入 history，避免请求体携带别名/展示名落库。
func (c *HistoryCtrl) canonicalEventNameForRow(ctx context.Context, eventID int64, fallback string) string {
	out := strings.TrimSpace(fallback)
	if eventID <= 0 {
		return out
	}
	events, err := c.Svc.ListEventOptions(ctx)
	if err != nil {
		return out
	}
	for i := range events {
		if events[i].Id == eventID {
			if n := strings.TrimSpace(events[i].Name); n != "" {
				return n
			}
			break
		}
	}
	return out
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
// Redis 事件选项缓存与 device-service 共用键，logo 可能为 OSS objectKey；对外 HTTP MUST 映射为 CDN 绝对 URL。
func (c *HistoryCtrl) EventOptions(ctx context.Context, req *v1.DeviceHistoryEventOptionsReq) (res *v1.DeviceHistoryEventOptionsRes, err error) {
	_ = req
	items, err := c.Svc.ListEventOptions(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryEventOptionsRes{List: eventlogo.MapEventsLogoCdn(ctx, items)}, nil
}

// Birthday 获取设备生日。
func (c *HistoryCtrl) Birthday(ctx context.Context, req *v1.DeviceHistoryBirthdayGetReq) (res *v1.DeviceHistoryBirthdayGetRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	babyName, birthday, sex, err := c.Svc.GetBirthday(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryBirthdayGetRes{BabyName: babyName, Birthday: birthday, Sex: sex}, nil
}

// BirthdaySave 保存设备生日。
func (c *HistoryCtrl) BirthdaySave(ctx context.Context, req *v1.DeviceHistoryBirthdaySaveReq) (res *v1.DeviceHistoryBirthdaySaveRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	sex := 0
	if req.Sex > 0 {
		sex = 1
	}
	if err := c.Svc.SaveBirthday(ctx, deviceNo, strings.TrimSpace(req.BabyName), req.Birthday, sex); err != nil {
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
	chatCtx := ctx
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		wxID := voice.ParseHeaderWxID(r.GetHeader(gatewayapp.HeaderInternalWxId))
		chatCtx = voice.WithVoiceWxID(ctx, wxID)
	}
	reply, err := c.Voice.TextChat(chatCtx, deviceNo, transcript)
	if err != nil {
		return nil, mapVoiceAIQuotaErr(err)
	}
	return &v1.DeviceHistoryChatRes{Reply: reply}, nil
}

// EventAdd 手动新增历史事件。
func (c *HistoryCtrl) EventAdd(ctx context.Context, req *v1.DeviceHistoryEventAddReq) (res *v1.DeviceHistoryEventAddRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	id, err := c.Svc.AddHistory(ctx, entity.History{
		DeviceNo:    deviceNo,
		EventId:     req.EventId,
		EventName:   c.canonicalEventNameForRow(ctx, req.EventId, req.EventName),
		EventNumber: int64(req.EventNumber),
		EventUnit:   strings.TrimSpace(req.EventUnit),
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
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
	err = c.Svc.UpdateHistory(ctx, mergeHistoryUpdateFromReq(ctx, c, req))
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryEventUpdateRes{}, nil
}

func mergeHistoryUpdateFromReq(ctx context.Context, c *HistoryCtrl, req *v1.DeviceHistoryEventUpdateReq) entity.History {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	item, err := histsvc.GetDeviceHistoryByID(ctx, req.Id, deviceNo)
	if err != nil {
		item = entity.History{Id: req.Id, DeviceNo: deviceNo}
	}
	item.EventId = req.EventId
	item.EventName = c.canonicalEventNameForRow(ctx, req.EventId, req.EventName)
	item.EventNumber = int64(req.EventNumber)
	item.EventUnit = strings.TrimSpace(req.EventUnit)
	item.StartTime = req.StartTime
	item.EndTime = req.EndTime
	item.Remark = strings.TrimSpace(req.Remark)
	if req.PostId != nil {
		item.PostId = *req.PostId
	}
	if req.MediaType != nil {
		item.MediaType = *req.MediaType
	}
	if req.ImageKeys != nil {
		item.ImageKeys = strings.Join(req.ImageKeys, ",")
	}
	if req.VideoKey != nil {
		item.VideoKey = strings.TrimSpace(*req.VideoKey)
	}
	return item
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

// Piece 区段内事件历史（GET /device/history/api/piece）。
func (c *HistoryCtrl) Piece(ctx context.Context, req *v1.DeviceHistoryPieceReq) (res *v1.DeviceHistoryPieceRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if req.EventId <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "eventId 无效")
	}
	list, err := histsvc.ListHistoryPiece(ctx, deviceNo, req.EventId, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryPieceRes{List: list}, nil
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
	updated, err := c.Svc.EndLatestHistoryIfMatch(ctx, deviceNo, req.EventId, req.EndTime, strings.TrimSpace(req.Remark))
	if err != nil {
		return nil, err
	}
	return &v1.DeviceHistoryEndLatestRes{Updated: updated}, nil
}
