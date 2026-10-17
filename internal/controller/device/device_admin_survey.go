package devicectrl

import (
	"context"

	v1 "hello/api/v1"
	device "hello/internal/services/device"
)

// SurveyEditor GET /device/admin/api/survey/editor
func (c *AdminCtrl) SurveyEditor(ctx context.Context, req *v1.DeviceAdminSurveyEditorReq) (res *v1.DeviceAdminSurveyEditorRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	ed, err := device.SurveyEditorGet(ctx, req.CampaignId)
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
	opts := make([]device.SurveyOptionInput, 0, len(req.Options))
	for _, opt := range req.Options {
		opts = append(opts, device.SurveyOptionInput{Id: opt.Id, Text: opt.Text, Enabled: opt.Enabled})
	}
	id, err := device.SurveyEditorSave(ctx, req.CampaignId, req.Enabled, req.StartAt, req.EndAt, req.Intro, opts)
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
	id, err := device.SurveyNewCampaign(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminSurveyNewRes{CampaignId: id}, nil
}

// SurveyResponses GET /device/admin/api/survey/responses
func (c *AdminCtrl) SurveyResponses(ctx context.Context, req *v1.DeviceAdminSurveyResponsesReq) (res *v1.DeviceAdminSurveyResponsesRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	page, err := device.SurveyResponses(ctx, req.CampaignId, req.Page, req.PageSize)
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
