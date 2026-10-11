package voicectrl

// voice_admin_predict_report.go：Hub 设备数据运维「预测上报」只读接口。
//
// 业务：运维在 history.html 核对 App 上报到 Redis 的 imminent 待办与进行中 ETA。
// 设计：鉴权同其它 voice Admin（X-Admin-Password，由 gateway 注入）；复用 Load*，
// 不改写 Redis、不投递 MQ、不推送。与 internal GET 字段对齐，供浏览器经 Admin JWT 访问。

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// VoiceAdminPredictReportCtrl Hub 预测上报只读（imminent + active-eta）。
type VoiceAdminPredictReportCtrl struct{}

// NewVoiceAdminPredictReportCtrl 构造管理端预测上报控制器。
func NewVoiceAdminPredictReportCtrl() *VoiceAdminPredictReportCtrl {
	return &VoiceAdminPredictReportCtrl{}
}

// requireAdmin 校验 X-Admin-Password（与 rule-gaps / intent-vectors 一致）。
func (c *VoiceAdminPredictReportCtrl) requireAdmin(ctx context.Context) error {
	_ = c
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	if !voice.VerifyVoiceAdminPassword(ctx, strings.TrimSpace(r.GetHeader("X-Admin-Password"))) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	return nil
}

// ImminentPendingGet GET /voice/admin/api/predict/imminent/pending?deviceNo=
// 返回 Redis 当前临近待办；缺键空 events；只读。
func (c *VoiceAdminPredictReportCtrl) ImminentPendingGet(ctx context.Context, req *v1.VoiceAdminPredictImminentPendingGetReq) (res *v1.VoiceAdminPredictImminentPendingGetRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	list, err := voice.LoadPredictImminentPending(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	// 映射完整字段（含可选 expect*），与 App PUT / internal GET 对齐。
	out := make([]v1.PredictImminentPendingEventItem, 0, len(list))
	for _, it := range list {
		out = append(out, v1.PredictImminentPendingEventItem{
			EventId:           it.EventId,
			NextAt:            it.NextAt,
			Title:             it.Title,
			ExpectKind:        it.ExpectKind,
			ExpectDurationSec: it.ExpectDurationSec,
			ExpectAmount:      it.ExpectAmount,
			ExpectUnit:        it.ExpectUnit,
			ExpectLeafTitle:   it.ExpectLeafTitle,
		})
	}
	return &v1.VoiceAdminPredictImminentPendingGetRes{Events: out}, nil
}

// ActiveEtaGet GET /voice/admin/api/predict/active-eta?deviceNo=
// 返回 Redis 当前进行中 ETA；缺键空 items；只读。
func (c *VoiceAdminPredictReportCtrl) ActiveEtaGet(ctx context.Context, req *v1.VoiceAdminPredictActiveEtaGetReq) (res *v1.VoiceAdminPredictActiveEtaGetRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	list, err := voice.LoadPredictActiveEta(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	out := make([]v1.PredictActiveEtaItem, 0, len(list))
	for _, it := range list {
		out = append(out, v1.PredictActiveEtaItem{
			RootEventId: it.RootEventId,
			LeafEventId: it.LeafEventId,
			HistoryId:   it.HistoryId,
			Title:       it.Title,
			StartAt:     it.StartAt,
			ExpectEndAt: it.ExpectEndAt,
		})
	}
	return &v1.VoiceAdminPredictActiveEtaGetRes{Items: out}, nil
}
