package controller

import (
	"context"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
)

// DeviceAppFeedbackCtrl App 反馈 API（JWT + X-Internal-Wx-Id）。
type DeviceAppFeedbackCtrl struct{}

// NewDeviceAppFeedbackCtrl 构造 App 反馈控制器。
func NewDeviceAppFeedbackCtrl() *DeviceAppFeedbackCtrl { return &DeviceAppFeedbackCtrl{} }

func feedbackTimeUnix(t *gtime.Time) int64 {
	if t == nil || t.IsZero() {
		return 0
	}
	return t.Timestamp()
}

func feedbackTimeUnixPtr(t *gtime.Time) *int64 {
	if t == nil || t.IsZero() {
		return nil
	}
	v := t.Timestamp()
	return &v
}

func toAppFeedbackItem(row deviceFeedbackRow) v1.DeviceAppFeedbackItem {
	item := v1.DeviceAppFeedbackItem{
		Id:        row.Id,
		Question:  row.Question,
		Status:    row.Status,
		CreatedAt: feedbackTimeUnix(row.CreatedAt),
	}
	if row.OfficialReply != "" {
		item.OfficialReply = row.OfficialReply
	}
	item.RepliedAt = feedbackTimeUnixPtr(row.RepliedAt)
	return item
}

type deviceFeedbackRow struct {
	Id            int64
	Question      string
	OfficialReply string
	Status        int
	CreatedAt     *gtime.Time
	RepliedAt     *gtime.Time
}

// List GET /device/app/api/feedback/list
func (c *DeviceAppFeedbackCtrl) List(ctx context.Context, req *v1.DeviceAppFeedbackListReq) (res *v1.DeviceAppFeedbackListRes, err error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	rows, err := device.DeviceAdmin().ListFeedbackByWxID(ctx, wxID)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAppFeedbackItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toAppFeedbackItem(deviceFeedbackRow{
			Id:            row.Id,
			Question:      row.Question,
			OfficialReply: row.OfficialReply,
			Status:        row.Status,
			CreatedAt:     row.CreatedAt,
			RepliedAt:     row.RepliedAt,
		}))
	}
	return &v1.DeviceAppFeedbackListRes{List: list}, nil
}

// Submit POST /device/app/api/feedback/submit
func (c *DeviceAppFeedbackCtrl) Submit(ctx context.Context, req *v1.DeviceAppFeedbackSubmitReq) (res *v1.DeviceAppFeedbackSubmitRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	row, err := device.DeviceAdmin().SubmitFeedback(ctx, wxID, req.Question)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	return &v1.DeviceAppFeedbackSubmitRes{
		Id:        row.Id,
		Question:  row.Question,
		Status:    row.Status,
		CreatedAt: feedbackTimeUnix(row.CreatedAt),
	}, nil
}
