package device

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	surveyIntroMaxRunes   = 500
	surveyOptionMaxRunes  = 200
	surveyOptionMaxRows   = 40
	surveyCommentMaxRunes = 1000
)

// SurveyOptionView App / Hub 共用的选项行。
type SurveyOptionView struct {
	Id      int64  `json:"id"`
	Text    string `json:"text"`
	Sort    int    `json:"sort"`
	Enabled bool   `json:"enabled"`
}

// SurveyCurrent App 当前是否展示。
type SurveyCurrent struct {
	Show       bool               `json:"show"`
	CampaignId int64              `json:"campaignId"`
	Intro      string             `json:"intro"`
	Options    []SurveyOptionView `json:"options"`
}

// SurveyEditor 运维编辑态。
type SurveyEditor struct {
	CampaignId int64                 `json:"campaignId"`
	Enabled    bool                  `json:"enabled"`
	StartAt    int64                 `json:"startAt"`
	EndAt      int64                 `json:"endAt"`
	Intro      string                `json:"intro"`
	Options    []SurveyOptionView    `json:"options"`
	Campaigns  []SurveyCampaignBrief `json:"campaigns"`
	IsLatest   bool                  `json:"isLatest"`
}

// SurveyCampaignBrief 活动下拉。
type SurveyCampaignBrief struct {
	Id      int64 `json:"id"`
	Enabled bool  `json:"enabled"`
	StartAt int64 `json:"startAt"`
	EndAt   int64 `json:"endAt"`
}

// SurveyOptionInput 保存时的一行。id=0 表示新增。
type SurveyOptionInput struct {
	Id      int64
	Text    string
	Enabled bool
}

// SurveyResponseItem Hub 提交行。
type SurveyResponseItem struct {
	WxId        int64    `json:"wxId"`
	BabyName    string   `json:"babyName"`
	DeviceNo    string   `json:"deviceNo"`
	Options     []string `json:"options"`
	Comment     string   `json:"comment"`
	SubmittedAt int64    `json:"submittedAt"`
}

// SurveyResponsePage 提交分页。
type SurveyResponsePage struct {
	List       []SurveyResponseItem `json:"list"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"pageSize"`
	CampaignId int64                `json:"campaignId"`
}

// SurveyCurrentForWx 读取最新活动对该用户是否展示。
// 无活动、关闭、不在 [startAt,endAt]、或已提交时 show=false。无启用选项仍可展示（允许只写意见）。
func SurveyCurrentForWx(ctx context.Context, wxID int64) (*SurveyCurrent, error) {
	out := &SurveyCurrent{Options: []SurveyOptionView{}}
	camp, err := latestSurveyCampaign(ctx)
	if err != nil || camp == nil {
		return out, err
	}
	out.CampaignId = camp.Id
	if !surveyCampaignOpen(camp, time.Now().Unix()) {
		return out, nil
	}
	done, err := surveyWxSubmitted(ctx, camp.Id, wxID)
	if err != nil || done {
		return out, err
	}
	opts, err := listSurveyOptions(ctx, camp.Id, true)
	if err != nil {
		return nil, err
	}
	out.Show = true
	out.Intro = camp.Intro
	out.Options = opts
	return out, nil
}

// SurveySubmit 写入提交。双空拒绝且不落库；重复提交保持单行并成功。
func SurveySubmit(ctx context.Context, wxID int64, deviceNo string, campaignID int64, optionIDs []int64, comment string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "未登录")
	}
	comment = strings.TrimSpace(comment)
	if utf8.RuneCountInString(comment) > surveyCommentMaxRunes {
		return gerror.NewCode(gcode.CodeInvalidParameter, "意见过长")
	}
	optionIDs = uniquePositiveIDs(optionIDs)
	if len(optionIDs) == 0 && comment == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "请至少选择一项或填写意见")
	}
	camp, err := latestSurveyCampaign(ctx)
	if err != nil {
		return err
	}
	if camp == nil || camp.Id != campaignID || !surveyCampaignOpen(camp, time.Now().Unix()) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "问卷当前不可提交")
	}
	done, err := surveyWxSubmitted(ctx, camp.Id, wxID)
	if err != nil {
		return err
	}
	if done {
		return nil
	}
	texts, err := surveyOptionTexts(ctx, camp.Id, optionIDs)
	if err != nil {
		return err
	}
	deviceNo = clipRunes(strings.TrimSpace(deviceNo), 64)
	babyName, err := ucgBabyNameByDeviceNo(ctx, deviceNo)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(texts)
	if err != nil {
		return err
	}
	babyName = clipRunes(babyName, 64)
	c := dao.SurveyResponse.Columns()
	_, err = dao.SurveyResponse.Ctx(ctx).Data(g.Map{
		c.CampaignId:     camp.Id,
		c.WxId:           wxID,
		c.DeviceNo:       deviceNo,
		c.BabyName:       babyName,
		c.OptionSnapshot: string(raw),
		c.Comment:        comment,
		c.SubmittedAt:    time.Now().Unix(),
	}).Insert()
	if err != nil {
		if surveyDupKey(err) {
			return nil
		}
		return gerror.WrapCode(gcode.CodeInternalError, err, "保存问卷失败")
	}
	return nil
}

// SurveyEditorGet 读取指定活动；campaignID<=0 时取最新。没有活动时返回空编辑态。
func SurveyEditorGet(ctx context.Context, campaignID int64) (*SurveyEditor, error) {
	briefs, latestID, err := listSurveyCampaignBriefs(ctx)
	if err != nil {
		return nil, err
	}
	out := &SurveyEditor{Options: []SurveyOptionView{}, Campaigns: briefs}
	if latestID == 0 {
		return out, nil
	}
	id := campaignID
	if id <= 0 {
		id = latestID
	}
	camp, err := surveyCampaignByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if camp == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "问卷活动不存在")
	}
	opts, err := listSurveyOptions(ctx, camp.Id, false)
	if err != nil {
		return nil, err
	}
	out.CampaignId = camp.Id
	out.Enabled = camp.Enabled == 1
	out.StartAt = camp.StartAt
	out.EndAt = camp.EndAt
	out.Intro = camp.Intro
	out.Options = opts
	out.IsLatest = camp.Id == latestID
	return out, nil
}

// SurveyEditorSave 保存活动开关、有效期、说明与选项行。起止无效则拒绝。
func SurveyEditorSave(ctx context.Context, campaignID int64, enabled bool, startAt, endAt int64, intro string, options []SurveyOptionInput) (int64, error) {
	intro = strings.TrimSpace(intro)
	if utf8.RuneCountInString(intro) > surveyIntroMaxRunes {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "说明过长")
	}
	if startAt <= 0 || endAt <= 0 || endAt < startAt {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "请设置有效的开始与结束时间")
	}
	if len(options) > surveyOptionMaxRows {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "选项过多")
	}
	for i := range options {
		options[i].Text = strings.TrimSpace(options[i].Text)
		if options[i].Text == "" || utf8.RuneCountInString(options[i].Text) > surveyOptionMaxRunes {
			return 0, gerror.NewCode(gcode.CodeInvalidParameter, "选项文案不能为空且需在 200 字以内")
		}
	}
	if campaignID <= 0 {
		id, err := insertSurveyCampaign(ctx, enabled, startAt, endAt, intro)
		if err != nil {
			return 0, err
		}
		if err := replaceSurveyOptions(ctx, id, options); err != nil {
			return 0, err
		}
		return id, nil
	}
	camp, err := surveyCampaignByID(ctx, campaignID)
	if err != nil {
		return 0, err
	}
	if camp == nil {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "问卷活动不存在")
	}
	c := dao.SurveyCampaign.Columns()
	en := 0
	if enabled {
		en = 1
	}
	_, err = dao.SurveyCampaign.Ctx(ctx).Where(c.Id, campaignID).Data(g.Map{
		c.Enabled: en,
		c.StartAt: startAt,
		c.EndAt:   endAt,
		c.Intro:   intro,
	}).Update()
	if err != nil {
		return 0, err
	}
	if err := replaceSurveyOptions(ctx, campaignID, options); err != nil {
		return 0, err
	}
	return campaignID, nil
}

// SurveyNewCampaign 开启新一轮：插入关闭的空活动，不复制旧提交。
func SurveyNewCampaign(ctx context.Context) (int64, error) {
	return insertSurveyCampaign(ctx, false, 0, 0, "")
}

// SurveyResponses 按活动分页列出提交。campaignID<=0 时用最新活动。
func SurveyResponses(ctx context.Context, campaignID int64, page, pageSize int) (*SurveyResponsePage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	if campaignID <= 0 {
		camp, err := latestSurveyCampaign(ctx)
		if err != nil {
			return nil, err
		}
		if camp == nil {
			return &SurveyResponsePage{List: []SurveyResponseItem{}, Page: page, PageSize: pageSize}, nil
		}
		campaignID = camp.Id
	}
	c := dao.SurveyResponse.Columns()
	m := dao.SurveyResponse.Ctx(ctx).Where(c.CampaignId, campaignID)
	total, err := m.Count()
	if err != nil {
		return nil, err
	}
	var rows []entity.SurveyResponse
	err = dao.SurveyResponse.Ctx(ctx).
		Where(c.CampaignId, campaignID).
		OrderDesc(c.SubmittedAt).
		Page(page, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	list := make([]SurveyResponseItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, SurveyResponseItem{
			WxId:        row.WxId,
			BabyName:    row.BabyName,
			DeviceNo:    row.DeviceNo,
			Options:     decodeOptionSnapshot(row.OptionSnapshot),
			Comment:     row.Comment,
			SubmittedAt: row.SubmittedAt,
		})
	}
	return &SurveyResponsePage{
		List: list, Total: total, Page: page, PageSize: pageSize, CampaignId: campaignID,
	}, nil
}

func surveyCampaignOpen(camp *entity.SurveyCampaign, now int64) bool {
	if camp == nil || camp.Enabled != 1 {
		return false
	}
	if camp.EndAt < camp.StartAt {
		return false
	}
	return now >= camp.StartAt && now <= camp.EndAt
}

func latestSurveyCampaign(ctx context.Context) (*entity.SurveyCampaign, error) {
	c := dao.SurveyCampaign.Columns()
	one, err := dao.SurveyCampaign.Ctx(ctx).OrderDesc(c.Id).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, nil
	}
	var row entity.SurveyCampaign
	if err := one.Struct(&row); err != nil {
		return nil, err
	}
	return &row, nil
}

func surveyCampaignByID(ctx context.Context, id int64) (*entity.SurveyCampaign, error) {
	c := dao.SurveyCampaign.Columns()
	one, err := dao.SurveyCampaign.Ctx(ctx).Where(c.Id, id).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, nil
	}
	var row entity.SurveyCampaign
	if err := one.Struct(&row); err != nil {
		return nil, err
	}
	return &row, nil
}

func surveyWxSubmitted(ctx context.Context, campaignID, wxID int64) (bool, error) {
	c := dao.SurveyResponse.Columns()
	one, err := dao.SurveyResponse.Ctx(ctx).
		Where(c.CampaignId, campaignID).
		Where(c.WxId, wxID).
		Fields(c.Id).
		One()
	if err != nil {
		return false, err
	}
	return !one.IsEmpty(), nil
}

func listSurveyOptions(ctx context.Context, campaignID int64, enabledOnly bool) ([]SurveyOptionView, error) {
	c := dao.SurveyOption.Columns()
	m := dao.SurveyOption.Ctx(ctx).Where(c.CampaignId, campaignID)
	if enabledOnly {
		m = m.Where(c.Enabled, 1)
	}
	var rows []entity.SurveyOption
	if err := m.OrderAsc(c.Sort).OrderAsc(c.Id).Scan(&rows); err != nil {
		return nil, err
	}
	out := make([]SurveyOptionView, 0, len(rows))
	for _, row := range rows {
		out = append(out, SurveyOptionView{
			Id: row.Id, Text: row.Text, Sort: row.Sort, Enabled: row.Enabled == 1,
		})
	}
	return out, nil
}

func surveyOptionTexts(ctx context.Context, campaignID int64, ids []int64) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	c := dao.SurveyOption.Columns()
	var rows []entity.SurveyOption
	err := dao.SurveyOption.Ctx(ctx).
		Where(c.CampaignId, campaignID).
		Where(c.Enabled, 1).
		WhereIn(c.Id, ids).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]string, len(rows))
	for _, row := range rows {
		byID[row.Id] = row.Text
	}
	texts := make([]string, 0, len(ids))
	for _, id := range ids {
		text, ok := byID[id]
		if !ok {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "包含无效或已关闭的选项")
		}
		texts = append(texts, text)
	}
	return texts, nil
}

func insertSurveyCampaign(ctx context.Context, enabled bool, startAt, endAt int64, intro string) (int64, error) {
	c := dao.SurveyCampaign.Columns()
	en := 0
	if enabled {
		en = 1
	}
	id, err := dao.SurveyCampaign.Ctx(ctx).Data(g.Map{
		c.Enabled:   en,
		c.StartAt:   startAt,
		c.EndAt:     endAt,
		c.Intro:     intro,
		c.CreatedAt: time.Now().Unix(),
	}).InsertAndGetId()
	return id, err
}

func replaceSurveyOptions(ctx context.Context, campaignID int64, options []SurveyOptionInput) error {
	return dao.SurveyCampaign.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		c := dao.SurveyOption.Columns()
		keep := make([]int64, 0, len(options))
		for i, opt := range options {
			en := 0
			if opt.Enabled {
				en = 1
			}
			if opt.Id > 0 {
				n, err := dao.SurveyOption.Ctx(ctx).
					Where(c.Id, opt.Id).
					Where(c.CampaignId, campaignID).
					Data(g.Map{c.Text: opt.Text, c.Sort: i, c.Enabled: en}).
					UpdateAndGetAffected()
				if err != nil {
					return err
				}
				if n == 0 {
					return gerror.NewCode(gcode.CodeInvalidParameter, "选项不属于该活动")
				}
				keep = append(keep, opt.Id)
				continue
			}
			id, err := dao.SurveyOption.Ctx(ctx).Data(g.Map{
				c.CampaignId: campaignID,
				c.Text:       opt.Text,
				c.Sort:       i,
				c.Enabled:    en,
			}).InsertAndGetId()
			if err != nil {
				return err
			}
			keep = append(keep, id)
		}
		del := dao.SurveyOption.Ctx(ctx).Where(c.CampaignId, campaignID)
		if len(keep) > 0 {
			del = del.WhereNotIn(c.Id, keep)
		}
		_, err := del.Delete()
		return err
	})
}

func listSurveyCampaignBriefs(ctx context.Context) ([]SurveyCampaignBrief, int64, error) {
	c := dao.SurveyCampaign.Columns()
	var rows []entity.SurveyCampaign
	err := dao.SurveyCampaign.Ctx(ctx).OrderDesc(c.Id).Limit(50).Scan(&rows)
	if err != nil {
		return nil, 0, err
	}
	out := make([]SurveyCampaignBrief, 0, len(rows))
	var latest int64
	for i, row := range rows {
		if i == 0 {
			latest = row.Id
		}
		out = append(out, SurveyCampaignBrief{
			Id: row.Id, Enabled: row.Enabled == 1, StartAt: row.StartAt, EndAt: row.EndAt,
		})
	}
	return out, latest, nil
}

func uniquePositiveIDs(ids []int64) []int64 {
	seen := map[int64]struct{}{}
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func decodeOptionSnapshot(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	var texts []string
	if err := json.Unmarshal([]byte(raw), &texts); err != nil {
		return []string{raw}
	}
	if texts == nil {
		return []string{}
	}
	return texts
}

func clipRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

func surveyDupKey(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "1062")
}
