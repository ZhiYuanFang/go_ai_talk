package device

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	surveyIntroMaxRunes   = 500
	surveyOptionMaxRunes  = 200
	surveyOptionMaxRows   = 40
	surveyCommentMaxRunes = 1000
	surveyResultMaxRunes  = 203
	surveyResultMaxCount  = 40
)

// surveyResultSuffixes 客户端拼在选项原文后的评价，无分隔符。
var surveyResultSuffixes = []string{"未使用", "不好用", "好用"}

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

// SurveyOptionInput 保留给历史保存入参；选项现已逐行写入，整表保存不再使用。
type SurveyOptionInput struct {
	Id      int64
	Text    string
	Enabled bool
}

// SurveyResponseItem Hub 提交行。
type SurveyResponseItem struct {
	Id          int64    `json:"id"`
	WxId        int64    `json:"wxId"`
	BabyName    string   `json:"babyName"`
	DeviceNo    string   `json:"deviceNo"`
	Options     []string `json:"options"`
	Comment     string   `json:"comment"`
	SubmittedAt int64    `json:"submittedAt"`
	Valid       bool     `json:"valid"`
}

// SurveyResponsePage 提交分页。
type SurveyResponsePage struct {
	List       []SurveyResponseItem `json:"list"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"pageSize"`
	CampaignId int64                `json:"campaignId"`
}

// SurveyCurrentForWx 读取唯一问卷对该用户是否展示。
// 关闭、不在 [startAt,endAt]、或已提交时 show=false。无有效文案的选项仍允许只写意见。
func SurveyCurrentForWx(ctx context.Context, wxID int64) (*SurveyCurrent, error) {
	out := &SurveyCurrent{Options: []SurveyOptionView{}}
	camp, err := EnsureSurveySingleCampaign(ctx)
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

// SurveySubmit 写入提交。结果必须是启用选项原文加评价后缀；双空拒绝且不落库。
// 已有有效提交则保持原行。无效行再交卷时用本次 optionResults 覆盖快照。
func SurveySubmit(ctx context.Context, wxID int64, deviceNo string, campaignID int64, optionResults []string, comment string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "未登录")
	}
	comment = strings.TrimSpace(comment)
	if utf8.RuneCountInString(comment) > surveyCommentMaxRunes {
		return gerror.NewCode(gcode.CodeInvalidParameter, "意见过长")
	}
	if len(optionResults) == 0 && comment == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "请至少选择一项或填写意见")
	}
	camp, err := EnsureSurveySingleCampaign(ctx)
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
	opts, err := listSurveyOptions(ctx, camp.Id, true)
	if err != nil {
		return err
	}
	texts, err := normalizeSurveyOptionResults(optionResults, opts)
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
	existing, err := surveyResponseByWx(ctx, camp.Id, wxID)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	if existing != nil {
		// 无效行再交卷：覆盖同一行并恢复有效。
		_, err = dao.SurveyResponse.Ctx(ctx).Where(c.Id, existing.Id).Data(g.Map{
			c.DeviceNo:       deviceNo,
			c.BabyName:       babyName,
			c.OptionSnapshot: string(raw),
			c.Comment:        comment,
			c.SubmittedAt:    now,
			c.Valid:          1,
		}).Update()
		return err
	}
	_, err = dao.SurveyResponse.Ctx(ctx).Data(g.Map{
		c.CampaignId:     camp.Id,
		c.WxId:           wxID,
		c.DeviceNo:       deviceNo,
		c.BabyName:       babyName,
		c.OptionSnapshot: string(raw),
		c.Comment:        comment,
		c.SubmittedAt:    now,
		c.Valid:          1,
	}).Insert()
	if err != nil {
		if surveyDupKey(err) {
			return nil
		}
		return gerror.WrapCode(gcode.CodeInternalError, err, "保存问卷失败")
	}
	return nil
}

// SurveyEditorGet 读取唯一问卷。没有活动时会先插入默认关闭的一行。
func SurveyEditorGet(ctx context.Context) (*SurveyEditor, error) {
	camp, err := EnsureSurveySingleCampaign(ctx)
	if err != nil {
		return nil, err
	}
	opts, err := listSurveyOptions(ctx, camp.Id, false)
	if err != nil {
		return nil, err
	}
	return &SurveyEditor{
		CampaignId: camp.Id,
		Enabled:    camp.Enabled == 1,
		StartAt:    camp.StartAt,
		EndAt:      camp.EndAt,
		Intro:      camp.Intro,
		Options:    opts,
		Campaigns:  []SurveyCampaignBrief{},
		IsLatest:   true,
	}, nil
}

// SurveyEditorSave 只保存开关、有效期和说明，不改选项行。起止无效则拒绝。
func SurveyEditorSave(ctx context.Context, enabled bool, startAt, endAt int64, intro string) (int64, error) {
	intro = strings.TrimSpace(intro)
	if utf8.RuneCountInString(intro) > surveyIntroMaxRunes {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "说明过长")
	}
	if startAt <= 0 || endAt <= 0 || endAt < startAt {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "请设置有效的开始与结束时间")
	}
	camp, err := EnsureSurveySingleCampaign(ctx)
	if err != nil {
		return 0, err
	}
	c := dao.SurveyCampaign.Columns()
	en := 0
	if enabled {
		en = 1
	}
	_, err = dao.SurveyCampaign.Ctx(ctx).Where(c.Id, camp.Id).Data(g.Map{
		c.Enabled: en,
		c.StartAt: startAt,
		c.EndAt:   endAt,
		c.Intro:   intro,
	}).Update()
	if err != nil {
		return 0, err
	}
	return camp.Id, nil
}

// SurveyAddOption 新增一行并立即落库。空文案允许暂存，App 不展示。
func SurveyAddOption(ctx context.Context, text string, enabled bool) (*SurveyOptionView, error) {
	text = strings.TrimSpace(text)
	if utf8.RuneCountInString(text) > surveyOptionMaxRunes {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "选项文案需在 200 字以内")
	}
	camp, err := EnsureSurveySingleCampaign(ctx)
	if err != nil {
		return nil, err
	}
	c := dao.SurveyOption.Columns()
	n, err := dao.SurveyOption.Ctx(ctx).Where(c.CampaignId, camp.Id).Count()
	if err != nil {
		return nil, err
	}
	if n >= surveyOptionMaxRows {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "选项过多")
	}
	en := 0
	if enabled {
		en = 1
	}
	id, err := dao.SurveyOption.Ctx(ctx).Data(g.Map{
		c.CampaignId: camp.Id,
		c.Text:       text,
		c.Sort:       n,
		c.Enabled:    en,
	}).InsertAndGetId()
	if err != nil {
		return nil, err
	}
	return &SurveyOptionView{Id: id, Text: text, Sort: n, Enabled: enabled}, nil
}

// SurveyUpdateOption 更新一行文案与是否启用。
func SurveyUpdateOption(ctx context.Context, id int64, text string, enabled bool) error {
	if id <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "选项不存在")
	}
	text = strings.TrimSpace(text)
	if utf8.RuneCountInString(text) > surveyOptionMaxRunes {
		return gerror.NewCode(gcode.CodeInvalidParameter, "选项文案需在 200 字以内")
	}
	camp, err := EnsureSurveySingleCampaign(ctx)
	if err != nil {
		return err
	}
	c := dao.SurveyOption.Columns()
	en := 0
	if enabled {
		en = 1
	}
	affected, err := dao.SurveyOption.Ctx(ctx).
		Where(c.Id, id).
		Where(c.CampaignId, camp.Id).
		Data(g.Map{c.Text: text, c.Enabled: en}).
		UpdateAndGetAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "选项不存在")
	}
	return nil
}

// SurveyDeleteOption 删除一行。已有提交快照不动。
func SurveyDeleteOption(ctx context.Context, id int64) error {
	if id <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "选项不存在")
	}
	camp, err := EnsureSurveySingleCampaign(ctx)
	if err != nil {
		return err
	}
	c := dao.SurveyOption.Columns()
	_, err = dao.SurveyOption.Ctx(ctx).Where(c.Id, id).Where(c.CampaignId, camp.Id).Delete()
	return err
}

// SurveyResponses 列出唯一问卷的全部提交。
func SurveyResponses(ctx context.Context, page, pageSize int) (*SurveyResponsePage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	camp, err := EnsureSurveySingleCampaign(ctx)
	if err != nil {
		return nil, err
	}
	campaignID := camp.Id
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
			Id:          row.Id,
			WxId:        row.WxId,
			BabyName:    row.BabyName,
			DeviceNo:    row.DeviceNo,
			Options:     decodeOptionSnapshot(row.OptionSnapshot),
			Comment:     row.Comment,
			SubmittedAt: row.SubmittedAt,
			Valid:       row.Valid != 0,
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

// EnsureSurveySingleCampaign 保证库里只有一份问卷。
// 没有则插入默认关闭的一行；多行时并入最小 id，同一 wx 只留较早的提交。
func EnsureSurveySingleCampaign(ctx context.Context) (*entity.SurveyCampaign, error) {
	c := dao.SurveyCampaign.Columns()
	var rows []entity.SurveyCampaign
	if err := dao.SurveyCampaign.Ctx(ctx).OrderAsc(c.Id).Scan(&rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		id, err := insertSurveyCampaign(ctx, false, 0, 0, "")
		if err != nil {
			return nil, err
		}
		return surveyCampaignByID(ctx, id)
	}
	keep := rows[0]
	for _, extra := range rows[1:] {
		if err := mergeSurveyCampaignInto(ctx, keep.Id, extra.Id); err != nil {
			return nil, err
		}
	}
	return surveyCampaignByID(ctx, keep.Id)
}

// mergeSurveyCampaignInto 把 extra 的选项和提交并入 keep，然后删除 extra。
func mergeSurveyCampaignInto(ctx context.Context, keepID, extraID int64) error {
	oc := dao.SurveyOption.Columns()
	if _, err := dao.SurveyOption.Ctx(ctx).Where(oc.CampaignId, extraID).Data(g.Map{oc.CampaignId: keepID}).Update(); err != nil {
		return err
	}
	rc := dao.SurveyResponse.Columns()
	var extras []entity.SurveyResponse
	if err := dao.SurveyResponse.Ctx(ctx).Where(rc.CampaignId, extraID).Scan(&extras); err != nil {
		return err
	}
	for _, ex := range extras {
		one, err := dao.SurveyResponse.Ctx(ctx).Where(rc.CampaignId, keepID).Where(rc.WxId, ex.WxId).One()
		if err != nil {
			return err
		}
		if one.IsEmpty() {
			if _, err := dao.SurveyResponse.Ctx(ctx).Where(rc.Id, ex.Id).Data(g.Map{rc.CampaignId: keepID}).Update(); err != nil {
				return err
			}
			continue
		}
		var kept entity.SurveyResponse
		if err := one.Struct(&kept); err != nil {
			return err
		}
		// 同一用户留较早的那次提交。
		if ex.SubmittedAt < kept.SubmittedAt {
			if _, err := dao.SurveyResponse.Ctx(ctx).Where(rc.Id, kept.Id).Delete(); err != nil {
				return err
			}
			if _, err := dao.SurveyResponse.Ctx(ctx).Where(rc.Id, ex.Id).Data(g.Map{rc.CampaignId: keepID}).Update(); err != nil {
				return err
			}
			continue
		}
		if _, err := dao.SurveyResponse.Ctx(ctx).Where(rc.Id, ex.Id).Delete(); err != nil {
			return err
		}
	}
	cc := dao.SurveyCampaign.Columns()
	_, err := dao.SurveyCampaign.Ctx(ctx).Where(cc.Id, extraID).Delete()
	return err
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
	row, err := surveyResponseByWx(ctx, campaignID, wxID)
	if err != nil || row == nil {
		return false, err
	}
	return row.Valid != 0, nil
}

func surveyResponseByWx(ctx context.Context, campaignID, wxID int64) (*entity.SurveyResponse, error) {
	c := dao.SurveyResponse.Columns()
	one, err := dao.SurveyResponse.Ctx(ctx).
		Where(c.CampaignId, campaignID).
		Where(c.WxId, wxID).
		One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, nil
	}
	var row entity.SurveyResponse
	if err := one.Struct(&row); err != nil {
		return nil, err
	}
	return &row, nil
}

// SurveySetResponseValid 把一条提交标为有效或无效。无效行仍保留。
func SurveySetResponseValid(ctx context.Context, id int64, valid bool) error {
	if id <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "提交不存在")
	}
	camp, err := EnsureSurveySingleCampaign(ctx)
	if err != nil {
		return err
	}
	c := dao.SurveyResponse.Columns()
	v := 0
	if valid {
		v = 1
	}
	n, err := dao.SurveyResponse.Ctx(ctx).
		Where(c.Id, id).
		Where(c.CampaignId, camp.Id).
		Data(g.Map{c.Valid: v}).
		UpdateAndGetAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "提交不存在")
	}
	return nil
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
		// 空文案只给 Hub 编辑，不发给 App。
		if enabledOnly && strings.TrimSpace(row.Text) == "" {
			continue
		}
		out = append(out, SurveyOptionView{
			Id: row.Id, Text: row.Text, Sort: row.Sort, Enabled: row.Enabled == 1,
		})
	}
	return out, nil
}

// normalizeSurveyOptionResults 校验客户端提交的「原文+评价」并保持提交顺序。
// 空串、重复、超长、对不上启用选项时整单拒绝。多个选项都能对上时仍接受。
func normalizeSurveyOptionResults(results []string, options []SurveyOptionView) ([]string, error) {
	if len(results) > surveyResultMaxCount {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "问卷结果过多")
	}
	texts := make([]string, 0, len(options))
	for _, opt := range options {
		text := strings.TrimSpace(opt.Text)
		if text == "" {
			continue
		}
		texts = append(texts, text)
	}
	seen := make(map[string]struct{}, len(results))
	out := make([]string, 0, len(results))
	for _, raw := range results {
		item := strings.TrimSpace(raw)
		if item == "" {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "问卷结果不能为空")
		}
		if utf8.RuneCountInString(item) > surveyResultMaxRunes {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "问卷结果过长")
		}
		if _, ok := seen[item]; ok {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "问卷结果重复")
		}
		if !surveyResultMatchesOption(item, texts) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "问卷结果与启用选项不匹配")
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out, nil
}

func surveyResultMatchesOption(result string, optionTexts []string) bool {
	for _, text := range optionTexts {
		for _, suffix := range surveyResultSuffixes {
			if result == text+suffix {
				return true
			}
		}
	}
	return false
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
