package devicectrl

import (
	"context"
	"errors"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeviceAppPredictProIdeaCtrl App 预测 Pro 想法墙 API（JWT；不查 Pro 权益）。
type DeviceAppPredictProIdeaCtrl struct{}

// NewDeviceAppPredictProIdeaCtrl 构造想法墙 App 控制器。
func NewDeviceAppPredictProIdeaCtrl() *DeviceAppPredictProIdeaCtrl {
	return &DeviceAppPredictProIdeaCtrl{}
}

func mapPredictProIdeaAppItem(it device.PredictProIdeaAppItem) v1.DeviceAppPredictProIdeaItem {
	return v1.DeviceAppPredictProIdeaItem{
		Id: it.Id, Text: it.Text, BabyName: it.BabyName,
		LikeCount: it.LikeCount, LikedByMe: it.LikedByMe, IsMine: it.IsMine, CreatedAt: it.CreatedAt,
	}
}

func mapPredictProIdeaBizErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, device.ErrPredictProIdeaNotFound):
		return gerror.NewCode(gcode.CodeNotFound, err.Error())
	case errors.Is(err, device.ErrPredictProIdeaQuota),
		errors.Is(err, device.ErrPredictProIdeaNoBind),
		errors.Is(err, device.ErrPredictProIdeaNoBabyName):
		return gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	default:
		if gerror.Code(err) != gcode.CodeNil {
			return err
		}
		return gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
}

// List GET /device/app/api/predict-pro-ideas/list
func (c *DeviceAppPredictProIdeaCtrl) List(ctx context.Context, req *v1.DeviceAppPredictProIdeaListReq) (*v1.DeviceAppPredictProIdeaListRes, error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	rows, err := device.DeviceAdmin().ListPredictProIdeasApp(ctx, wxID)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAppPredictProIdeaItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, mapPredictProIdeaAppItem(row))
	}
	return &v1.DeviceAppPredictProIdeaListRes{List: list}, nil
}

// Submit POST /device/app/api/predict-pro-ideas/submit
func (c *DeviceAppPredictProIdeaCtrl) Submit(ctx context.Context, req *v1.DeviceAppPredictProIdeaSubmitReq) (*v1.DeviceAppPredictProIdeaSubmitRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	item, err := device.DeviceAdmin().SubmitPredictProIdea(ctx, wxID, req.Text)
	if err != nil {
		return nil, mapPredictProIdeaBizErr(err)
	}
	mapped := mapPredictProIdeaAppItem(*item)
	return &v1.DeviceAppPredictProIdeaSubmitRes{
		Id: mapped.Id, Text: mapped.Text, BabyName: mapped.BabyName,
		LikeCount: mapped.LikeCount, LikedByMe: mapped.LikedByMe, IsMine: mapped.IsMine, CreatedAt: mapped.CreatedAt,
	}, nil
}

// Like POST /device/app/api/predict-pro-ideas/like
func (c *DeviceAppPredictProIdeaCtrl) Like(ctx context.Context, req *v1.DeviceAppPredictProIdeaLikeReq) (*v1.DeviceAppPredictProIdeaLikeRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	out, err := device.DeviceAdmin().LikePredictProIdea(ctx, wxID, req.Id)
	if err != nil {
		return nil, mapPredictProIdeaBizErr(err)
	}
	return &v1.DeviceAppPredictProIdeaLikeRes{LikeCount: out.LikeCount, LikedByMe: out.LikedByMe}, nil
}

// Unlike POST /device/app/api/predict-pro-ideas/unlike
func (c *DeviceAppPredictProIdeaCtrl) Unlike(ctx context.Context, req *v1.DeviceAppPredictProIdeaUnlikeReq) (*v1.DeviceAppPredictProIdeaLikeRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	out, err := device.DeviceAdmin().UnlikePredictProIdea(ctx, wxID, req.Id)
	if err != nil {
		return nil, mapPredictProIdeaBizErr(err)
	}
	return &v1.DeviceAppPredictProIdeaLikeRes{LikeCount: out.LikeCount, LikedByMe: out.LikedByMe}, nil
}
