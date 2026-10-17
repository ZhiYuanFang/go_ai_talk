package devicectrl

import (
	"context"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// SurveyEditor GET /device/admin/api/survey/editor
func (c *AdminCtrl) SurveyEditor(ctx context.Context, _ *v1.DeviceAdminSurveyEditorReq) (res *v1.DeviceAdminSurveyEditorRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	ed, err := device.SurveyEditorGet(ctx)
	if err != nil {
		return nil, err
	}
	return adminSurveyEditorRes(ed), nil
}

// SurveySave POST /device/admin/api/survey/save
func (c *AdminCtrl) SurveySave(ctx context.Context, req *v1.DeviceAdminSurveySaveReq) (res *v1.DeviceAdminSurveySaveRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id, err := device.SurveyEditorSave(ctx, req.Enabled, req.StartAt, req.EndAt, req.Intro)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminSurveySaveRes{CampaignId: id}, nil
}

// SurveyNew POST /device/admin/api/survey/new
func (c *AdminCtrl) SurveyNew(ctx context.Context, _ *v1.DeviceAdminSurveyNewReq) (res *v1.DeviceAdminSurveyNewRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, gerror.NewCode(gcode.CodeInvalidOperation, "问卷只有一份，不能开启新一轮")
}

// SurveyResponses GET /device/admin/api/survey/responses
func (c *AdminCtrl) SurveyResponses(ctx context.Context, req *v1.DeviceAdminSurveyResponsesReq) (res *v1.DeviceAdminSurveyResponsesRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	page, err := device.SurveyResponses(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminSurveyResponseItem, 0, len(page.List))
	for _, row := range page.List {
		opts := row.Options
		if opts == nil {
			opts = []string{}
		}
		list = append(list, v1.DeviceAdminSurveyResponseItem{
			WxId: row.WxId, BabyName: row.BabyName, DeviceNo: row.DeviceNo,
			Options: opts, Comment: row.Comment, SubmittedAt: row.SubmittedAt,
		})
	}
	return &v1.DeviceAdminSurveyResponsesRes{
		List: list, Total: page.Total, Page: page.Page, PageSize: page.PageSize, CampaignId: page.CampaignId,
	}, nil
}

// SurveyOptionAdd POST /device/admin/api/survey/option/add
func (c *AdminCtrl) SurveyOptionAdd(ctx context.Context, req *v1.DeviceAdminSurveyOptionAddReq) (res *v1.DeviceAdminSurveyOptionAddRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	row, err := device.SurveyAddOption(ctx, req.Text, req.Enabled)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminSurveyOptionAddRes{Id: row.Id, Text: row.Text, Sort: row.Sort, Enabled: row.Enabled}, nil
}

// SurveyOptionUpdate POST /device/admin/api/survey/option/update
func (c *AdminCtrl) SurveyOptionUpdate(ctx context.Context, req *v1.DeviceAdminSurveyOptionUpdateReq) (res *v1.DeviceAdminSurveyOptionUpdateRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := device.SurveyUpdateOption(ctx, req.Id, req.Text, req.Enabled); err != nil {
		return nil, err
	}
	return &v1.DeviceAdminSurveyOptionUpdateRes{}, nil
}

// SurveyOptionDelete POST /device/admin/api/survey/option/delete
func (c *AdminCtrl) SurveyOptionDelete(ctx context.Context, req *v1.DeviceAdminSurveyOptionDeleteReq) (res *v1.DeviceAdminSurveyOptionDeleteRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := device.SurveyDeleteOption(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeviceAdminSurveyOptionDeleteRes{}, nil
}

func adminSurveyEditorRes(ed *device.SurveyEditor) *v1.DeviceAdminSurveyEditorRes {
	opts := make([]v1.DeviceAdminSurveyOption, 0, len(ed.Options))
	for _, opt := range ed.Options {
		opts = append(opts, v1.DeviceAdminSurveyOption{Id: opt.Id, Text: opt.Text, Sort: opt.Sort, Enabled: opt.Enabled})
	}
	briefs := make([]v1.DeviceAdminSurveyCampaignBrief, 0, len(ed.Campaigns))
	for _, b := range ed.Campaigns {
		briefs = append(briefs, v1.DeviceAdminSurveyCampaignBrief{Id: b.Id, Enabled: b.Enabled, StartAt: b.StartAt, EndAt: b.EndAt})
	}
	return &v1.DeviceAdminSurveyEditorRes{
		CampaignId: ed.CampaignId, Enabled: ed.Enabled, StartAt: ed.StartAt, EndAt: ed.EndAt,
		Intro: ed.Intro, Options: opts, Campaigns: briefs, IsLatest: ed.IsLatest,
	}
}
